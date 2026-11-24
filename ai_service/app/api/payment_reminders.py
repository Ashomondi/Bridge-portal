from fastapi import APIRouter, HTTPException

router = APIRouter()


@router.post("/payment-reminders/draft", status_code=501)
async def draft_payment_reminder():
    raise HTTPException(
        status_code=501,
        detail={
            "code": "NOT_IMPLEMENTED",
            "message": "Payment reminders are a Release 2 feature",
        },
    )
