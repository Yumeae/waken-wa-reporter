//go:build windows && !waken_powershell_media

package media

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media/control"
)

const (
	hresultRPCAlreadyInitialized = 0x80010106
	asyncWaitTimeout             = 8 * time.Second
	asyncPollInterval            = 8 * time.Millisecond
)

var comOnce sync.Once

func ensureCOM() error {
	var initErr error
	comOnce.Do(func() {
		err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
		if err != nil {
			var oe *ole.OleError
			if errors.As(err, &oe) && oe.Code() == hresultRPCAlreadyInitialized {
				return
			}
			initErr = err
		}
	})
	return initErr
}

func waitAsyncCompleted(op *foundation.IAsyncOperation, timeout time.Duration) error {
	if op == nil {
		return errors.New("media: nil async operation")
	}
	itf := op.MustQueryInterface(ole.NewGUID(foundation.GUIDIAsyncInfo))
	defer itf.Release()
	asyncInfo := (*foundation.IAsyncInfo)(unsafe.Pointer(itf))
	deadline := time.Now().Add(timeout)
	for {
		st, err := asyncInfo.GetStatus()
		if err != nil {
			return err
		}
		switch st {
		case foundation.AsyncStatusCompleted:
			return nil
		case foundation.AsyncStatusError:
			return fmt.Errorf("media: async error status")
		case foundation.AsyncStatusCanceled:
			return fmt.Errorf("media: async canceled")
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("media: async wait timeout")
		}
		time.Sleep(asyncPollInterval)
	}
}

func getNowPlayingWinRT() (Info, error) {
	if err := ensureCOM(); err != nil {
		return Info{}, err
	}

	mgrAsync, err := control.GlobalSystemMediaTransportControlsSessionManagerRequestAsync()
	if err != nil {
		return Info{}, err
	}
	if err := waitAsyncCompleted(mgrAsync, asyncWaitTimeout); err != nil {
		mgrAsync.Release()
		return Info{}, err
	}
	mgrPtr, err := mgrAsync.GetResults()
	mgrAsync.Release()
	if err != nil {
		return Info{}, err
	}
	if mgrPtr == nil {
		return Info{}, ErrNoMedia
	}
	mgr := (*control.GlobalSystemMediaTransportControlsSessionManager)(unsafe.Pointer(mgrPtr))
	defer mgr.Release()

	session, err := mgr.GetCurrentSession()
	if err != nil {
		return Info{}, err
	}
	if session == nil {
		return Info{}, ErrNoMedia
	}
	defer session.Release()

	var out Info
	if sourceAppID, err := session.GetSourceAppUserModelId(); err == nil {
		out.SourceAppID = sourceAppID
	}

	propAsync, err := session.TryGetMediaPropertiesAsync()
	if err != nil {
		return Info{}, err
	}
	if err := waitAsyncCompleted(propAsync, asyncWaitTimeout); err != nil {
		propAsync.Release()
		return Info{}, err
	}
	propPtr, err := propAsync.GetResults()
	propAsync.Release()
	if err != nil {
		return Info{}, err
	}
	if propPtr == nil {
		return Info{}, ErrNoMedia
	}
	props := (*control.GlobalSystemMediaTransportControlsSessionMediaProperties)(unsafe.Pointer(propPtr))
	defer props.Release()

	if t, err := props.GetTitle(); err == nil {
		out.Title = t
	}
	if a, err := props.GetArtist(); err == nil {
		out.Artist = a
	}
	if al, err := props.GetAlbumTitle(); err == nil {
		out.Album = al
	}
	if out.IsEmpty() {
		return Info{}, ErrNoMedia
	}
	return out, nil
}

// GetNowPlaying reads the current Global System Media Transport session (Windows 10+).
// It prefers the native WinRT path and falls back to PowerShell when native interop fails
// or returns no active media metadata.
func GetNowPlaying() (Info, error) {
	info, err := getNowPlayingWinRT()
	if err == nil {
		return info, nil
	}
	psInfo, psErr := getNowPlayingViaPowerShell()
	if psErr == nil {
		return psInfo, nil
	}
	if errors.Is(err, ErrNoMedia) {
		return Info{}, psErr
	}
	return Info{}, fmt.Errorf("media: native=%v; powershell=%v", err, psErr)
}
