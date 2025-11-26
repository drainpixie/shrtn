use actix_web::{HttpResponse, ResponseError, http::StatusCode};
use rand::{Rng, distr::Alphanumeric, rng};
use serde::Serialize;
use thiserror::Error;

pub fn random_alphanumeric(len: usize) -> String {
	rng().sample_iter(&Alphanumeric).take(len).map(char::from).collect()
}

#[derive(Serialize)]
pub struct APIResponse<T> {
	pub success: bool,
	pub data: Option<T>,
	pub message: Option<String>,
}

#[derive(Error, Debug)]
pub enum APIError {
	#[error("invalid target url")]
	InvalidURL,

	#[error("short code already exists")]
	AlreadyExists,

	#[error("not found")]
	NotFound,

	#[error("unauthorized")]
	Unauthorized,

	#[error("database error {0}")]
	DatabaseError(String),

	#[error("database error {0}")]
	SqlxError(#[from] sqlx::Error),

	#[error("template rendering error {0}")]
	TemplateError(#[from] askama::Error),
}

pub type APIResult<T = HttpResponse> = Result<T, APIError>;

impl ResponseError for APIError {
	fn status_code(&self) -> StatusCode {
		match self {
			Self::NotFound => StatusCode::NOT_FOUND,
			Self::InvalidURL => StatusCode::BAD_REQUEST,
			Self::AlreadyExists => StatusCode::CONFLICT,
			Self::Unauthorized => StatusCode::UNAUTHORIZED,
			Self::SqlxError(_) => StatusCode::INTERNAL_SERVER_ERROR,
			Self::DatabaseError(_) => StatusCode::INTERNAL_SERVER_ERROR,
			Self::TemplateError(_) => StatusCode::INTERNAL_SERVER_ERROR,
		}
	}
}
