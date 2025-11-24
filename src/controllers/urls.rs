use actix_web::{HttpResponse, get, post, web};
use log::{error, info};
use serde::Deserialize;
use url::Url;

use crate::{
	services::urls::UrlService,
	utils::{APIError, APIResponse, APIResult, generate_short_code},
};

#[derive(Deserialize)]
pub struct AddRequest {
	pub short: String,
	pub target: String,
}

#[get("")]
pub async fn list(service: web::Data<UrlService>) -> APIResult {
	let urls = service.list().await.map_err(|e| {
		error!("failed to fetch urls {:?}", e);
		APIError::DatabaseError("failed to fetch urls".to_string())
	})?;

	Ok(HttpResponse::Ok().json(APIResponse {
		success: true,
		data: Some(urls),
		message: None,
	}))
}

#[post("")]
pub async fn add(
	service: web::Data<UrlService>,
	data: web::Json<AddRequest>,
) -> APIResult {
	let short = if data.short.is_empty() {
		generate_short_code(6)
	} else {
		data.short.clone()
	};

	Url::parse(&data.target).map_err(|_| APIError::InvalidURL)?;
	info!("adding new url short={}, target={}", short, data.target);

	let exists = service.exists(&short).await.map_err(|e| {
		error!("failed to check if short exists: {:?}", e);
		APIError::DatabaseError("failed to check existence".to_string())
	})?;

	if exists {
		return Err(APIError::AlreadyExists);
	}

	service.add(&short, &data.target).await.map_err(|e| {
		error!("db insert error {}", e);
		APIError::DatabaseError("failed to insert".to_string())
	})?;

	Ok(HttpResponse::Ok().json(APIResponse {
		success: true,
		data: Some(serde_json::json!({
			"short": short,
			"target": data.target,
		})),
		message: None,
	}))
}

pub fn configure(cfg: &mut web::ServiceConfig) {
	cfg.service(web::scope("/api/urls").service(list).service(add));
}
