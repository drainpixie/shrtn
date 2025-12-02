use std::collections::HashMap;

use actix_web::{HttpResponse, delete, get, post, web};
use log::{error, info};
use serde::Deserialize;
use url::Url;

use crate::{
	services::urls::UrlService,
	utils::{APIError, APIResponse, APIResult, random_alphanumeric},
};

#[derive(Deserialize)]
pub struct AddRequest {
	pub short: String,
	pub target: String,
}

#[get("{slug}")]
pub async fn get(
	slug: web::Path<String>,
	service: web::Data<UrlService>,
) -> APIResult {
	let slug = slug.into_inner();
	let url = service.get(&slug).await.map_err(|_| APIError::NotFound)?;

	Ok(HttpResponse::Ok().json(APIResponse {
		success: true,
		data: Some(url),
		message: None,
	}))
}

#[get("")]
pub async fn list(service: web::Data<UrlService>) -> APIResult {
	// TODO: guard this with an 'admin' token or similar

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
		random_alphanumeric(6)
	} else {
		data.short.clone()
	};

	Url::parse(&data.target).map_err(|_| APIError::InvalidURL)?;
	info!("adding new url short={} target={}", short, data.target);

	let exists = service.exists(&short).await.map_err(|e| {
		error!("failed to check if short exists: {:?}", e);
		APIError::DatabaseError("failed to check existence".to_string())
	})?;

	if exists {
		return Err(APIError::AlreadyExists);
	}

	let token = random_alphanumeric(32);

	service.add(&short, &data.target, &token).await.map_err(|e| {
		error!("db insert error {}", e);
		APIError::DatabaseError("failed to insert".to_string())
	})?;

	Ok(HttpResponse::Ok().json(APIResponse {
		success: true,
		data: Some(serde_json::json!({
		"short": short,
		"target": data.target,
		"token": token,
		})),
		message: None,
	}))
}

#[delete("/{slug}")]
pub async fn delete(
	service: web::Data<UrlService>,
	slug: web::Path<String>,
	query: web::Query<HashMap<String, String>>,
) -> APIResult {
	let slug = slug.into_inner();
	let token = match query.get("token") {
		Some(t) => t,
		None => return Err(APIError::Unauthorized),
	};

	let deleted = service.delete(&slug, token).await.map_err(|e| {
		error!("db delete error {}", e);
		APIError::DatabaseError("failed to delete".to_string())
	})?;

	if !deleted {
		return Err(APIError::NotFound);
	}

	Ok(HttpResponse::Ok().json(APIResponse::<()> {
		success: true,
		data: None,
		message: Some("url deleted successfully".to_string()),
	}))
}

pub fn configure(cfg: &mut web::ServiceConfig) {
	cfg.service(
		web::scope("/api/urls")
			.service(delete)
			.service(list)
			.service(add)
			.service(get),
	);
}
