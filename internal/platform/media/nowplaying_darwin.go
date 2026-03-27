//go:build darwin && cgo

package media

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -F/System/Library/PrivateFrameworks -framework MediaRemote

#import <Foundation/Foundation.h>
#import <dispatch/dispatch.h>
#import <stdlib.h>

void MRMediaRemoteGetNowPlayingInfo(dispatch_queue_t queue, void (^handler)(NSDictionary *info));

static int mediaRemoteFetch(char **title, char **artist, char **album) {
    dispatch_semaphore_t sem = dispatch_semaphore_create(0);
    __block int ok = 0;
    MRMediaRemoteGetNowPlayingInfo(dispatch_get_global_queue(DISPATCH_QUEUE_PRIORITY_DEFAULT, 0), ^(NSDictionary *info) {
        if (!info) {
            dispatch_semaphore_signal(sem);
            return;
        }
        NSString *t = info[@"kMRMediaRemoteNowPlayingInfoTitle"];
        NSString *a = info[@"kMRMediaRemoteNowPlayingInfoArtist"];
        NSString *al = info[@"kMRMediaRemoteNowPlayingInfoAlbum"];
        if (t && [t length] > 0) {
            *title = strdup([t UTF8String]);
            ok = 1;
        }
        if (a && [a length] > 0) {
            *artist = strdup([a UTF8String]);
            ok = 1;
        }
        if (al && [al length] > 0) {
            *album = strdup([al UTF8String]);
            ok = 1;
        }
        dispatch_semaphore_signal(sem);
    });
    dispatch_semaphore_wait(sem, dispatch_time(DISPATCH_TIME_NOW, (int64_t)(2 * NSEC_PER_SEC)));
    return ok;
}
*/
import "C"

import (
	"unsafe"
)

// GetNowPlaying uses the private MediaRemote framework. This may break across macOS updates.
func GetNowPlaying() (Info, error) {
	var ct, ca, cal *C.char
	defer func() {
		if ct != nil {
			C.free(unsafe.Pointer(ct))
		}
		if ca != nil {
			C.free(unsafe.Pointer(ca))
		}
		if cal != nil {
			C.free(unsafe.Pointer(cal))
		}
	}()
	ok := C.mediaRemoteFetch(&ct, &ca, &cal)
	if ok == 0 {
		return Info{}, ErrNoMedia
	}
	var out Info
	if ct != nil {
		out.Title = C.GoString(ct)
	}
	if ca != nil {
		out.Artist = C.GoString(ca)
	}
	if cal != nil {
		out.Album = C.GoString(cal)
	}
	if out.IsEmpty() {
		return Info{}, ErrNoMedia
	}
	return out, nil
}
