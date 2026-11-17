from fastapi import APIRouter, HTTPException

router = APIRouter()


@router.post("/reconciliation/suggest", status_code=501)
async def suggest_reconciliation():
    raise HTTPException(
        status_code=501,
        detail={
            "code": "NOT_IMPLEMENTED",
            "message": "M-Pesa reconciliation is a Release 2 feature",
        },
    )
