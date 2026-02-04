from moddingway_api.routes.banforms_routes import router as banform_router
from moddingway_api.routes.banneduser_routes import router as banneduser_router
from moddingway_api.routes.mod_routes import router as mod_router
from moddingway_api.routes.user_routes import router as user_router

__all__ = ["banform_router", "banneduser_router", "mod_router", "user_router"]
