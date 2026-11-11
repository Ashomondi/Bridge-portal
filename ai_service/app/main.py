from contextlib import asynccontextmanager

import httpx
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from starlette.exceptions import HTTPException as StarletteHTTPException

from .api import (
    blog_safety,
    parent_assistant,
    payment_reminders,
    reconciliation,
    report_drafter,
    syllabus_mapper,
)
from .state import SessionLimitError


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with httpx.AsyncClient() as client:
        app.state.http_client = client
        yield


app = FastAPI(title="Ikonex Academy AI Service", version="1.0.0", lifespan=lifespan)

API_PREFIX = "/api/v1/ai"

app.include_router(parent_assistant.router, prefix=API_PREFIX)
app.include_router(report_drafter.router, prefix=API_PREFIX)
app.include_router(syllabus_mapper.router, prefix=API_PREFIX)
app.include_router(blog_safety.router, prefix=API_PREFIX)
app.include_router(reconciliation.router, prefix=API_PREFIX)
app.include_router(payment_reminders.router, prefix=API_PREFIX)


@app.get("/healthz")
async def healthz():
    return {"status": "ok"}


@app.exception_handler(SessionLimitError)
async def session_limit_handler(request: Request, exc: SessionLimitError):
    return JSONResponse(
        status_code=429,
        content={
            "status": "error",
            "error": {"code": exc.code, "message": "Session limit exceeded"},
            "data": {"reply": exc.reply},
        },
    )


@app.exception_handler(StarletteHTTPException)
async def http_exception_handler(request: Request, exc: StarletteHTTPException):
    if isinstance(exc.detail, dict):
        error = exc.detail
    else:
        error = {"code": "HTTP_ERROR", "message": str(exc.detail)}
    return JSONResponse(status_code=exc.status_code, content={"status": "error", "error": error})
