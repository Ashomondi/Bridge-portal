from fastapi import APIRouter, HTTPException

router = APIRouter()


@router.post("/syllabus/map", status_code=501)
async def map_syllabus():
    raise HTTPException(
        status_code=501,
        detail={"code": "NOT_IMPLEMENTED", "message": "Syllabus mapping is a Release 2 feature"},
    )
