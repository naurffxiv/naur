import logging
from datetime import UTC, datetime

from fastapi import APIRouter, HTTPException
from fastapi_pagination import Page

from moddingway.database import banforms_database, users_database
from moddingway.database.models import BanForm
from moddingway.settings import get_settings
from moddingway_api.routes.banneduser_routes import unban_user
from moddingway_api.schemas.ban_form_schema import FormRequest, UpdateRequest
from moddingway_api.utils.paginate import paginate, parse_pagination_params

router = APIRouter(prefix="/banforms")
settings = get_settings()
authHeader = {"Authorization": f"Bot {settings.discord_token}"}
logger = logging.getLogger(__name__)


@router.get("")
async def get_ban_forms() -> Page[BanForm]:
    page, size = parse_pagination_params()
    limit = size
    offset = (page - 1) * size

    db_form_list = banforms_database.get_ban_forms(limit, offset)
    total_count = banforms_database.get_form_count()

    return paginate(db_form_list, length_function=lambda _: total_count)


@router.get("/{form_id}")
async def get_ban_form_by_id(form_id: int) -> BanForm | None:
    db_form = banforms_database.get_ban_form(form_id)

    if not db_form:
        raise HTTPException(status_code=404, detail="Form not found")

    form = BanForm(
        form_id=db_form.form_id,
        user_id=db_form.user_id,
        reason=db_form.reason,
        approval=db_form.approval,
        approved_by=db_form.approved_by,
        submission_timestamp=db_form.submission_timestamp,
    )
    return form


@router.post("")
async def submit_form(request: FormRequest):
    db_user = users_database.get_user(int(request.user_id))

    if db_user:
        if not db_user.is_banned:
            return {"detail": f"User {request.user_id} is not currently banned"}

        form = BanForm(
            user_id=db_user.user_id,
            reason=request.reason,
            submission_timestamp=datetime.now(UTC),
        )

        result = banforms_database.add_form(form)

        if result:
            # TODO: logging to moddingway
            return {"detail": f"User {request.user_id}'s form has been submitted"}

        return {
            "detail": "There was an issue with the form submission, please try again"
        }


@router.patch("")
async def update_form(request: UpdateRequest):
    db_form = banforms_database.get_ban_form(int(request.form_id))
    if not db_form:
        raise HTTPException(status_code=404, detail="Form not found")

    result = banforms_database.update_form(
        request.form_id, request.approval, request.approver_id
    )
    if result is not None and result[0]:
        db_user = banforms_database.get_user_from_form(result[1])
        if db_user:
            try:
                await unban_user(str(db_user))
            except Exception as e:
                logger.error(
                    f"Error unbanning user {db_user} via banform approval: {e}"
                )

    return {"detail": f"Form {request.form_id} has been updated"}
