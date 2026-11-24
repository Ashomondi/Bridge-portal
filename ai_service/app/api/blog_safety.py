from fastapi import APIRouter, HTTPException

router = APIRouter()


@router.post("/blog-safety-check", status_code=501)
async def blog_safety_check():
    raise HTTPException(
        status_code=501,
        detail={"code": "NOT_IMPLEMENTED", "message": "Blog safety check is a Release 2 feature"},
    )
