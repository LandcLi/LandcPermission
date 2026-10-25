package internal

import (
	_ "github.com/LandcLi/LandcPermission/internal/dao_impl/permission"
	_ "github.com/LandcLi/LandcPermission/internal/dao_impl/role"
	_ "github.com/LandcLi/LandcPermission/internal/dao_impl/user_permission"
	_ "github.com/LandcLi/LandcPermission/internal/dao_impl/user_role"

	_ "github.com/LandcLi/LandcPermission/internal/service_impl/permission"
	_ "github.com/LandcLi/LandcPermission/internal/service_impl/role"

	_ "github.com/LandcLi/LandcPermission/internal/controller/permission"
	_ "github.com/LandcLi/LandcPermission/internal/controller/role"
)
