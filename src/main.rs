mod controllers;
mod database;
mod services;
mod utils;

use std::env;

use actix_web::{
	App,
	HttpServer,
	middleware::{Compress, NormalizePath, TrailingSlash},
	web::{self, Data},
};
use dotenvy::dotenv;
use sqlx::SqlitePool;

use crate::{database::initialise_database, services::urls::UrlService};

#[actix_web::main]
async fn main() -> std::io::Result<()> {
	dotenv().ok();
	env_logger::init();

	let host = env::var("HOST").unwrap_or_else(|_| "127.0.0.1".to_string());
	let port = env::var("PORT").unwrap_or_else(|_| "8080".to_string());
	let database_url = env::var("DATABASE_URL")
		.unwrap_or_else(|_| "sqlite://local.db".to_string());

	let addr = format!("{}:{}", host, port);

	let pool = SqlitePool::connect(&database_url)
		.await
		.expect("failed to connect to database");

	initialise_database(&pool) //
		.await
		.expect("failed to initialise database");

	let url_service = UrlService::new(pool.clone());

	log::info!("starting server at {}", addr);

	HttpServer::new(move || {
		App::new()
			.app_data(Data::new(url_service.clone()))
			.wrap(Compress::default())
			.wrap(NormalizePath::new(TrailingSlash::Trim))
			.app_data(web::Data::new(pool.clone()))
			.configure(controllers::urls::configure)
			.configure(controllers::templates::configure)
	})
	.bind(addr)?
	.run()
	.await
}
