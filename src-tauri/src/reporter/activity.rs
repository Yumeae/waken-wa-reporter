use serde::{Deserialize, Serialize};
use reqwest::Client;

#[derive(Debug, Serialize)]
pub struct ReportRequest {
    #[serde(rename = "generatedHashKey")]
    pub generated_hash_key: String,
    pub device: String,
    pub device_name: String,
    pub device_type: String,
    pub process_name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub process_title: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub battery_level: Option<i64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub is_charging: Option<bool>,
    pub push_mode: String,
    pub metadata: serde_json::Value,
}

#[derive(Debug, Deserialize)]
struct ApiResponse {
    #[allow(dead_code)]
    success: Option<bool>,
    message: Option<String>,
    #[serde(rename = "approvalUrl")]
    approval_url: Option<String>,
}

pub async fn post_activity(
    base_url: &str,
    token: &str,
    req: ReportRequest,
    bypass_proxy: bool,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let client = if bypass_proxy {
        Client::builder().no_proxy().build()?
    } else {
        Client::new()
    };

    let url = format!("{}/api/activity", base_url.trim_end_matches('/'));
    let resp = client
        .post(&url)
        .bearer_auth(token)
        .json(&req)
        .timeout(std::time::Duration::from_secs(30))
        .send()
        .await?;

    let status = resp.status();
    let body: ApiResponse = resp.json().await.unwrap_or(ApiResponse {
        success: None,
        message: None,
        approval_url: None,
    });

    match status.as_u16() {
        200 | 201 => Ok(()),
        202 => Err(format!(
            "设备待审核：{}",
            body.approval_url.as_deref().unwrap_or("(无链接)")
        )
        .into()),
        401 => Err("Token 无效或已过期".into()),
        400 => Err(format!(
            "请求参数错误：{}",
            body.message.as_deref().unwrap_or("unknown")
        )
        .into()),
        _ => Err(format!("HTTP {}", status).into()),
    }
}
