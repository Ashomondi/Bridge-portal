from tests.helpers import make_headers


def _payload(mode="term_comment", reason=None, subject=None):
    return {
        "mode": mode,
        "student_id": "stu-1",
        "term_id": "term-1",
        "teacher_reason": reason,
        "subject_name": subject,
    }


async def test_term_comment_strips_forbidden_terms(client, settings, fake_llm):
    fake_llm.text = (
        "Brian ranked first in class and may have ADHD. He shows strong effort this term."
    )
    response = await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(),
        headers=make_headers(settings, role="teacher", channel="in_app"),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    draft = data["draft_comment"]
    assert "ranked" not in draft.lower()
    assert "adhd" not in draft.lower()
    assert data["requires_teacher_approval"] is True
    assert draft


async def test_sanitizer_catches_positioning_and_hyphen_variants(client, settings, fake_llm):
    fake_llm.text = (
        "Brian was positioning at the top-performing end. He shows strong effort this term."
    )
    response = await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(),
        headers=make_headers(settings, role="teacher", channel="in_app"),
    )
    draft = response.json()["data"]["draft_comment"]
    assert "positioning" not in draft.lower()
    assert "top" not in draft.lower()
    assert "effort" in draft.lower()


async def test_rerelease_note_from_teacher_reason(client, settings):
    response = await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(
            mode="rerelease_note",
            reason="Mathematics paper 2 mark was transposed from 68 to 86",
            subject="Mathematics",
        ),
        headers=make_headers(settings, role="teacher", channel="in_app"),
    )
    assert response.status_code == 200
    data = response.json()["data"]
    assert data["draft_comment"]
    assert data["requires_teacher_approval"] is True


async def test_rerelease_requires_teacher_reason(client, settings):
    response = await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(mode="rerelease_note", subject="Mathematics"),
        headers=make_headers(settings, role="teacher", channel="in_app"),
    )
    assert response.status_code == 400


async def test_parent_role_cannot_draft(client, settings):
    response = await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(),
        headers=make_headers(settings, role="parent", channel="in_app"),
    )
    assert response.status_code == 403


async def test_report_draft_audit_written(client, settings, fake_backend):
    await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(),
        headers=make_headers(settings, role="teacher", channel="in_app"),
    )
    assert fake_backend.audit_calls
    assert fake_backend.audit_calls[0].feature == "report_drafter"


async def test_unsafe_fallback_is_sanitized(client, settings, fake_llm):
    fake_llm.text = ""
    response = await client.post(
        "/api/v1/ai/report-comments/draft",
        json=_payload(
            mode="rerelease_note",
            reason="The learner was ranked first and has ADHD",
            subject="Mathematics",
        ),
        headers=make_headers(settings, role="teacher", channel="in_app"),
    )
    assert response.status_code == 200
    draft = response.json()["data"]["draft_comment"]
    assert draft
    assert "ranked" not in draft.lower()
    assert "adhd" not in draft.lower()
