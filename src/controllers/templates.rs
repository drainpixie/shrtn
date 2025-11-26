use actix_web::{HttpResponse, get, web};
use askama::Template;

use crate::{services::urls::UrlService, utils::APIResult};

#[derive(Template)]
#[template(path = "index.html")]
struct IndexTemplate;

#[derive(Template)]
#[template(path = "404.html")]
struct NotFoundTemplate;

#[derive(Template)]
#[template(path = "info.html")]
struct InfoTemplate;

fn render_template<T: Template>(template: &T) -> APIResult {
	let html = template.render()?;
	Ok(HttpResponse::Ok().content_type("text/html").body(html))
}

#[get("/info")]
pub async fn info() -> APIResult {
	render_template(&InfoTemplate {})
}

#[get("/{slug:.*}")]
pub async fn index(
	slug: web::Path<String>,
	service: web::Data<UrlService>,
) -> APIResult {
	let slug = slug.into_inner();
	if slug.is_empty() {
		return render_template(&IndexTemplate {});
	}

	match service.get(&slug).await? {
		Some(url) => {
			service.click(&slug).await?;

			Ok(HttpResponse::Found()
				.append_header(("Location", url.target))
				.finish())
		}
		None => render_template(&NotFoundTemplate {}),
	}
}

pub fn configure(cfg: &mut web::ServiceConfig) {
	cfg.service(info).service(index);
}
