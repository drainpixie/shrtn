use serde::{Deserialize, Serialize};
use sqlx::{FromRow, SqlitePool, query};

#[derive(FromRow, Serialize, Deserialize)]
pub struct Url {
	pub id: i64,
	pub short: String,
	pub target: String,
}

#[derive(FromRow, Serialize, Deserialize)]
pub struct User {
	pub id: i64,
	pub username: String,
	pub password_hash: String,
}

pub async fn initialise_database(pool: &SqlitePool) -> Result<(), sqlx::Error> {
	let create_urls_table = r#"
    CREATE TABLE IF NOT EXISTS urls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        short TEXT NOT NULL UNIQUE,
        target TEXT NOT NULL
    );
    "#;

	let create_users_table = r#"
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL
    );
    "#;

	query(create_urls_table).execute(pool).await?;
	query(create_users_table).execute(pool).await?;

	Ok(())
}
