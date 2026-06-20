#!/usr/bin/env python3
"""Collect recent complaint reviews from RestaurantGuru.

The script crawls a handful of large city pages, extracts restaurant URLs,
fetches the paginated review JSON for each restaurant, filters to complaint-like
reviews from 2025 onward, and writes a compact research corpus.
"""

from __future__ import annotations

import calendar
import csv
import datetime as dt
import html
import json
import re
from collections import Counter, defaultdict
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


BASE = "https://restaurantguru.com.br"
OUT_DIR = Path("research/restaurant-complaints-2025")
CURRENT_DATE = dt.date(2026, 6, 20)
MIN_DATE = dt.date(2025, 1, 1)
SSL_CONTEXT = __import__("ssl")._create_unverified_context()

CITY_PAGES = [
    "Sao-Paulo",
    "Rio-de-Janeiro",
    "Belo-Horizonte",
    "Curitiba",
    "Porto-Alegre",
]

MAX_CITY_PAGES = 2
MAX_RESTAURANTS_PER_CITY = 40
MAX_REVIEW_PAGES_PER_RESTAURANT = 4
TARGET_COMPLAINTS = 1000

HARD_NEGATIVE_CUES = [
    "ruim",
    "péssim",
    "pessim",
    "horr",
    "terrív",
    "terriv",
    "decepcion",
    "insatisfa",
    "problema",
    "falha",
    "erro",
    "demora",
    "demorou",
    "demorado",
    "lento",
    "espera",
    "atras",
    "frio",
    "gelado",
    "cru",
    "queimad",
    "seco",
    "sem sabor",
    "insosso",
    "caro",
    "preço alto",
    "preco alto",
    "abusiv",
    "cobran",
    "cobrado",
    "sujo",
    "limpeza",
    "higiene",
    "odor",
    "cheiro ruim",
    "pedido errado",
    "veio errado",
    "faltou",
    "bagun",
    "pequena porção",
    "porção pequena",
    "pouca quantidade",
    "nunca mais",
    "não volto",
    "nao volto",
]

NEGATIVE_KEYWORDS = {
    "service": [
        "atendimento ruim",
        "atendimento péssim",
        "atendimento pessim",
        "atendimento lento",
        "garçom ruim",
        "garcom ruim",
        "staff rude",
        "antipático",
        "antipatico",
        "mal-educ",
        "rude",
    ],
    "delay": [
        "demora",
        "demorou",
        "demorado",
        "lento",
        "espera",
        "atras",
        "fila",
    ],
    "food_quality": [
        "ruim",
        "péssim",
        "pessim",
        "frio",
        "gelado",
        "cru",
        "queimad",
        "seco",
        "sem sabor",
        "insosso",
        "borrach",
        "mal coz",
    ],
    "order_accuracy": [
        "pedido errado",
        "veio errado",
        "errado",
        "faltou",
        "esquecer",
        "trocaram",
        "bagun",
        "marmita",
    ],
    "cleanliness": [
        "sujo",
        "limpeza",
        "higiene",
        "banheiro",
        "cheiro",
        "odor",
    ],
    "price_value": [
        "caro",
        "preço alto",
        "preco alto",
        "overpriced",
        "abusiv",
        "custo",
    ],
    "portion": [
        "pouco",
        "pequena",
        "porcion",
        "porção",
        "quantidade",
    ],
    "reservation_queue": [
        "reserva",
        "lotad",
        "mesa",
        "fila",
        "esperando",
    ],
    "billing": [
        "conta",
        "cobran",
        "cobrado",
        "fatura",
        "taxa",
    ],
    "delivery": [
        "delivery",
        "entrega",
        "embalagem",
        "marmita",
        "ifood",
        "retirada",
    ],
}


def fetch(url: str, referer: str | None = None) -> str:
    headers = {
        "User-Agent": "Mozilla/5.0",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    }
    if referer:
        headers["Referer"] = referer
    req = Request(url, headers=headers)
    with urlopen(req, timeout=30, context=SSL_CONTEXT) as resp:
        return resp.read().decode("utf-8", "ignore")


def fetch_json(url: str, referer: str) -> dict:
    headers = {
        "User-Agent": "Mozilla/5.0",
        "Accept": "application/json,text/plain,*/*",
        "X-Requested-With": "XMLHttpRequest",
        "Referer": referer,
    }
    req = Request(url, headers=headers)
    with urlopen(req, timeout=30, context=SSL_CONTEXT) as resp:
        return json.loads(resp.read().decode("utf-8", "ignore"))


def strip_tags(text: str) -> str:
    text = re.sub(r"<[^>]+>", " ", text)
    text = html.unescape(text)
    text = re.sub(r"\s+", " ", text)
    return text.strip()


def clean_review_body(text: str) -> str:
    markers = [
        "response from the owner",
        "a response from the owner",
        "resposta do proprietário",
        "resposta do proprietario",
        "resposta do dono",
    ]
    lowered = text.lower()
    cut_at = len(text)
    for marker in markers:
        idx = lowered.find(marker)
        if idx != -1:
            cut_at = min(cut_at, idx)
    cleaned = text[:cut_at].strip()
    cleaned = re.sub(r"\b[A-Z]\s*Response from the owner.*$", "", cleaned, flags=re.I | re.S).strip()
    return cleaned


def parse_city_restaurants(city: str, max_city_pages: int, max_per_city: int) -> list[str]:
    seen: set[str] = set()
    restaurants: list[str] = []
    for page in range(1, max_city_pages + 1):
        url = f"{BASE}/{city}" if page == 1 else f"{BASE}/{city}/{page}"
        html_text = fetch(url)
        links = re.findall(r'href=\"(https://restaurantguru\.com\.br/[^\"]+)\"', html_text)
        for href in links:
            if not href.startswith(f"{BASE}/"):
                continue
            if href == f"{BASE}/{city}" or href.startswith(f"{BASE}/{city}/"):
                continue
            if any(part in href for part in ("/menu", "/reviews", "/guides", "/businessLanding")):
                continue
            if href in seen:
                continue
            seen.add(href)
            restaurants.append(href)
            if len(restaurants) >= max_per_city:
                return restaurants
    return restaurants


def parse_review_page(page_html: str) -> list[dict]:
    cards = re.findall(
        r'(<div\s+data-id=\"[^\"]+\"[\s\S]*?<div class=\"text\">[\s\S]*?</div>\s*</div>\s*</div>\s*</div>)',
        page_html,
    )
    reviews: list[dict] = []
    for card in cards:
        score_m = re.search(r'data-score=\"(\d+)\"', card)
        author_m = re.search(r'<a class=\"user_info__name\"[^>]*>(.*?)</a>', card, re.S)
        date_m = re.search(r'<span class=\"grey\">([^<]+)</span>', card)
        text_m = re.search(r'<div class=\"text\">([\s\S]*?)</div>\s*</div>\s*</div>', card)
        if not (score_m and author_m and date_m and text_m):
            continue
        review_text = clean_review_body(strip_tags(text_m.group(1)))
        reviews.append(
            {
                "score": int(score_m.group(1)),
                "author": strip_tags(author_m.group(1)),
                "date_text": strip_tags(date_m.group(1)),
                "text": review_text,
            }
        )
    return reviews


def word_to_int(text: str) -> int | None:
    text = text.lower()
    if re.search(r"\b(?:um|uma)\b", text):
        return 1
    m = re.search(r"\b(\d+)\b", text)
    if m:
        return int(m.group(1))
    return None


def subtract_months(date: dt.date, months: int) -> dt.date:
    year = date.year
    month = date.month - months
    while month <= 0:
        month += 12
        year -= 1
    day = min(date.day, calendar.monthrange(year, month)[1])
    return dt.date(year, month, day)


def approximate_date(date_text: str) -> dt.date | None:
    s = date_text.lower()
    if "hoje" in s:
        return CURRENT_DATE
    if "ontem" in s:
        return CURRENT_DATE - dt.timedelta(days=1)
    week_m = re.search(r"\b(\d+|um|uma)\s+semana", s)
    day_m = re.search(r"\b(\d+|um|uma)\s+dia", s)
    month_m = re.search(r"\b(\d+|um|uma)\s+m[eê]s(?:es)?", s)
    year_m = re.search(r"\b(\d+|um|uma)\s+ano(?:s)?", s)
    years = 0
    months = 0
    if year_m:
        years = word_to_int(year_m.group(0)) or 1
    if month_m:
        months = word_to_int(month_m.group(0)) or 1
    if years or months:
        date = dt.date(CURRENT_DATE.year - years, CURRENT_DATE.month, CURRENT_DATE.day)
        if months:
            date = subtract_months(date, months)
        return date
    if week_m:
        n = word_to_int(week_m.group(0)) or 1
        return CURRENT_DATE - dt.timedelta(days=7 * n)
    if day_m:
        n = word_to_int(day_m.group(0)) or 1
        return CURRENT_DATE - dt.timedelta(days=n)
    return None


def complaint_tags(text: str, score: int) -> list[str]:
    tags: list[str] = []
    lowered = text.lower()
    for tag, needles in NEGATIVE_KEYWORDS.items():
        if any(needle in lowered for needle in needles):
            tags.append(tag)
    if any(term in lowered for term in ("atendimento", "service", "garçom", "garcom", "staff")) and any(
        cue in lowered for cue in ("ruim", "péssim", "pessim", "rude", "antipatic", "mal-educ", "demora", "demor", "lento", "espera", "atras")
    ):
        if "service" not in tags:
            tags.append("service")
    if score <= 3 and "low_rating" not in tags:
        tags.append("low_rating")
    return tags


def looks_like_complaint(text: str, score: int) -> bool:
    cleaned = strip_tags(text)
    if not cleaned or len(cleaned.split()) < 4:
        return False
    if "price per person" in cleaned.lower():
        return False
    lowered = cleaned.lower()
    return score <= 3 or any(cue in lowered for cue in HARD_NEGATIVE_CUES)


def restaurant_name_from_url(url: str) -> str:
    slug = url.rstrip("/").split("/")[-1]
    slug = slug.replace("-", " ")
    return slug.title()


def fetch_reviews_for_restaurant(restaurant_url: str, max_pages: int) -> list[dict]:
    out: list[dict] = []
    for page in range(1, max_pages + 1):
        url = f"{restaurant_url}/reviews" if page == 1 else f"{restaurant_url}/reviews/{page}"
        try:
            payload = fetch_json(url, referer=restaurant_url)
        except (HTTPError, URLError):
            break
        reviews = parse_review_page(payload.get("html", ""))
        if not reviews:
            break
        for review in reviews:
            approx = approximate_date(review["date_text"])
            if approx is not None and approx < MIN_DATE:
                continue
            review["approx_date"] = approx.isoformat() if approx else ""
            review["restaurant_url"] = restaurant_url
            review["restaurant_name"] = restaurant_name_from_url(restaurant_url)
            review["review_page"] = page
            review["complaint_tags"] = complaint_tags(review["text"], review["score"])
            if looks_like_complaint(review["text"], review["score"]):
                out.append(review)
        # Stop once the oldest review on the page is clearly pre-2025.
        if reviews:
            last_date = approximate_date(reviews[-1]["date_text"])
            if last_date is not None and last_date < MIN_DATE:
                break
    return out


def main() -> int:
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    all_rows: list[dict] = []
    restaurant_sources: list[dict] = []

    for city in CITY_PAGES:
        restaurants = parse_city_restaurants(city, MAX_CITY_PAGES, MAX_RESTAURANTS_PER_CITY)
        for restaurant_url in restaurants:
            if len(all_rows) >= TARGET_COMPLAINTS:
                break
            complaints = fetch_reviews_for_restaurant(restaurant_url, MAX_REVIEW_PAGES_PER_RESTAURANT)
            if not complaints:
                continue
            restaurant_sources.append(
                {
                    "city": city,
                    "restaurant_url": restaurant_url,
                    "restaurant_name": restaurant_name_from_url(restaurant_url),
                    "complaint_count": len(complaints),
                }
            )
            all_rows.extend(complaints)
        if len(all_rows) >= TARGET_COMPLAINTS:
            break

    all_rows = all_rows[:TARGET_COMPLAINTS]
    out_jsonl = OUT_DIR / "complaints.jsonl"
    out_csv = OUT_DIR / "complaints.csv"
    out_sources = OUT_DIR / "sources.csv"
    out_summary = OUT_DIR / "summary.md"

    with out_jsonl.open("w", encoding="utf-8") as f:
        for row in all_rows:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")

    with out_csv.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(
            f,
            fieldnames=[
                "restaurant_name",
                "restaurant_url",
                "review_page",
                "score",
                "author",
                "date_text",
                "approx_date",
                "complaint_tags",
                "text",
            ],
        )
        writer.writeheader()
        for row in all_rows:
            writer.writerow(
                {
                    "restaurant_name": row["restaurant_name"],
                    "restaurant_url": row["restaurant_url"],
                    "review_page": row["review_page"],
                    "score": row["score"],
                    "author": row["author"],
                    "date_text": row["date_text"],
                    "approx_date": row["approx_date"],
                    "complaint_tags": "|".join(row["complaint_tags"]),
                    "text": row["text"],
                }
            )

    with out_sources.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=["city", "restaurant_name", "restaurant_url", "complaint_count"])
        writer.writeheader()
        for row in restaurant_sources:
            writer.writerow(row)

    tag_counts = Counter()
    rating_counts = Counter()
    yearly_counts = Counter()
    sample_quotes = defaultdict(list)
    for row in all_rows:
        rating_counts[row["score"]] += 1
        for tag in row["complaint_tags"]:
            tag_counts[tag] += 1
        if row["approx_date"]:
            yearly_counts[row["approx_date"][:4]] += 1
        for tag in row["complaint_tags"][:2]:
            if len(sample_quotes[tag]) < 3:
                sample_quotes[tag].append(row["text"][:180])

    summary_lines = [
        "# Resumo de reclamações de restaurante",
        "",
        f"- Fonte principal: RestaurantGuru (`{BASE}`)",
        f"- Recorte: reviews de 2025 em diante",
        f"- Total de avaliações com reclamação: {len(all_rows)}",
        f"- Restaurantes consultados: {len({r['restaurant_url'] for r in restaurant_sources})}",
        "",
        "## Temas mais comuns",
    ]
    for tag, count in tag_counts.most_common(10):
        summary_lines.append(f"- {tag}: {count}")
    summary_lines.append("")
    summary_lines.append("## Distribuição por nota")
    for score, count in sorted(rating_counts.items()):
        summary_lines.append(f"- {score} estrela(s): {count}")
    summary_lines.append("")
    summary_lines.append("## Distribuição por ano aproximado")
    for year, count in sorted(yearly_counts.items()):
        summary_lines.append(f"- {year}: {count}")
    summary_lines.append("")
    summary_lines.append("## Exemplos curtos por tema")
    for tag, quotes in list(sample_quotes.items())[:8]:
        summary_lines.append(f"- {tag}:")
        for quote in quotes:
            summary_lines.append(f"  - {quote}")

    out_summary.write_text("\n".join(summary_lines) + "\n", encoding="utf-8")
    print(f"Wrote {len(all_rows)} complaint reviews to {OUT_DIR}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
