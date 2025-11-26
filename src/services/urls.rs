use sqlx::SqlitePool;

use crate::database::Url;

#[derive(Clone)]
pub struct UrlService {
	pub pool: SqlitePool,
}

impl UrlService {
	pub fn new(pool: SqlitePool) -> Self {
		Self { pool }
	}

	pub async fn list(&self) -> Result<Vec<Url>, sqlx::Error> {
		sqlx::query_as::<_, Url>(
			"SELECT id, clicks, short, target, created_at, token FROM urls",
		)
		.fetch_all(&self.pool)
		.await
	}

	pub async fn exists(&self, short: &str) -> Result<bool, sqlx::Error> {
		let exists: Option<(i64,)> =
			sqlx::query_as("SELECT 1 FROM urls WHERE short = ? LIMIT 1")
				.bind(short)
				.fetch_optional(&self.pool)
				.await?;

		Ok(exists.is_some())
	}

	pub async fn get(&self, short: &str) -> Result<Option<Url>, sqlx::Error> {
		let url: Option<Url> = sqlx::query_as(
			"SELECT id, clicks, short, target, created_at, token FROM urls WHERE short = ?",
		)
		.bind(short)
		.fetch_optional(&self.pool)
		.await?;

		Ok(url)
	}

	pub async fn click(&self, short: &str) -> Result<(), sqlx::Error> {
		sqlx::query("UPDATE urls SET clicks = clicks + 1 WHERE short = ?")
			.bind(short)
			.execute(&self.pool)
			.await?;

		Ok(())
	}

	pub async fn add(
		&self,
		short: &str,
		target: &str,
		token: &str,
	) -> Result<(), sqlx::Error> {
		sqlx::query("INSERT INTO urls (short, target, token) VALUES (?, ?, ?)")
			.bind(short)
			.bind(target)
			.bind(token)
			.execute(&self.pool)
			.await?;

		Ok(())
	}
}
