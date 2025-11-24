use serde::{Deserialize, Serialize};
use sqlx::{FromRow, SqlitePool, query};

#[derive(FromRow, Serialize, Deserialize)]
pub struct Url {
	pub id: i64,
	pub clicks: i64,
	pub short: String,
	pub target: String,
	pub created_at: String,
	pub delete_token: String,
}

pub async fn initialise_database(pool: &SqlitePool) -> Result<(), sqlx::Error> {
	query(
		r#"
    CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    clicks INTEGER NOT NULL DEFAULT 0,
    short TEXT NOT NULL UNIQUE,
    target TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delete_token TEXT NOT NULL
    );
    "#,
	)
	.execute(pool)
	.await?;

	Ok(())
}
