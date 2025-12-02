use actix_web::{HttpResponse, get, web};
use askama::Template;

use crate::{database::Url, services::urls::UrlService, utils::APIResult};

#[derive(Template)]
#[template(path = "index.html")]
struct IndexTemplate;

#[derive(Template)]
#[template(path = "404.html")]
struct NotFoundTemplate;

#[derive(Template)]
#[template(path = "info.html")]
struct InfoTemplate;

#[derive(Template)]
#[template(path = "ctrl.html")]
struct ControlTemplate;

fn render_template<T: Template>(template: &T) -> APIResult {
	let html = template.render()?;
	Ok(HttpResponse::Ok().content_type("text/html").body(html))
}

#[get("/info")]
pub async fn info() -> APIResult {
	render_template(&InfoTemplate {})
}

#[get("/ctrl")]
pub async fn modify() -> APIResult {
	render_template(&ControlTemplate {})
}

#[get("/")]
pub async fn home() -> APIResult {
	render_template(&IndexTemplate {})
}

#[get("/{slug}")]
pub async fn redirect(
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
	cfg.service(home).service(info).service(modify).service(redirect);
}
