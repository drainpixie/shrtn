use actix_web::{HttpResponse, Responder, get, web};
use askama::Template;

use crate::utils::APIResult;

#[derive(Template)]
#[template(path = "index.html")]
struct IndexTemplate;

fn render_template<T: Template>(template: &T) -> APIResult {
	let html = template.render()?;
	Ok(HttpResponse::Ok().content_type("text/html").body(html))
}

#[get("/")]
pub async fn index() -> impl Responder {
	render_template(&IndexTemplate {})
}

pub fn configure(cfg: &mut web::ServiceConfig) {
	cfg.service(index);
}
