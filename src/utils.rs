use actix_web::{HttpResponse, ResponseError, http::StatusCode};
use rand::{Rng, distr::Alphanumeric, rng};
use serde::Serialize;
use thiserror::Error;

pub fn generate_short_code(len: usize) -> String {
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
	InvalidUrl,

	#[error("short code already exists")]
	AlreadyExists,

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
			Self::InvalidUrl => StatusCode::BAD_REQUEST,
			Self::AlreadyExists => StatusCode::CONFLICT,
			Self::SqlxError(_) => StatusCode::INTERNAL_SERVER_ERROR,
			Self::DatabaseError(_) => StatusCode::INTERNAL_SERVER_ERROR,
			Self::TemplateError(_) => StatusCode::INTERNAL_SERVER_ERROR,
		}
	}

	fn error_response(&self) -> HttpResponse {
		let message = Some(match self {
			Self::InvalidUrl => "invalid target URL",
			Self::AlreadyExists => "short code already exists",
			Self::SqlxError(_) => "database error",
			Self::DatabaseError(_) => "database error",
			Self::TemplateError(_) => "template rendering error",
		});

		HttpResponse::build(self.status_code()).json(APIResponse::<()> {
			success: false,
			data: None,
			message: message.map(String::from),
		})
	}
}
