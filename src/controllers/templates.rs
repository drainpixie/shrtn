use actix_web::{HttpResponse, get, web};
use askama::Template;

use crate::{services::urls::UrlService, utils::APIResult};

#[derive(Template)]
#[template(path = "index.html")]
struct IndexTemplate;

fn render_template<T: Template>(template: &T) -> APIResult {
	let html = template.render()?;
	Ok(HttpResponse::Ok().content_type("text/html").body(html))
}

#[get("/{slug:.*}")]
pub async fn index(
	slug: web::Path<String>,
	url: web::Data<UrlService>,
) -> APIResult {
	let slug = slug.into_inner();
	if slug.is_empty() {
		return render_template(&IndexTemplate {});
	}

	match url.get(&slug).await? {
		Some(url) => Ok(HttpResponse::Found()
			.append_header(("Location", url.target))
			.finish()),
		None => Ok(HttpResponse::NotFound().body("404 Not Found")),
	}
}

pub fn configure(cfg: &mut web::ServiceConfig) {
	cfg.service(index);
}
