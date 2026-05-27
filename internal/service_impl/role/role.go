package role

import (
	"fmt"

	"github.com/LandcLi/LandcPermission/dao"
	"github.com/LandcLi/LandcPermission/model"
	"github.com/LandcLi/LandcPermission/service"
)

func init() {
	service.RegisterRoleService(&roleServiceImpl{})
}

type roleServiceImpl struct{}

func (s *roleServiceImpl) CreateRole(name, code, description string) (*model.Role, error) {
	// Check if code already exists
	existing, _ := dao.GetRoleDAO().GetByCode(code)
	if existing != nil {
		return nil, fmt.Errorf("role code '%s' already exists", code)
	}

	role := &model.Role{
		Name:        name,
		Code:        code,
		Description: description,
		Status:      1,
	}
	if err := dao.GetRoleDAO().Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleServiceImpl) UpdateRole(id uint, name, code, description string, status *int) error {
	role, err := dao.GetRoleDAO().GetByID(id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if name != "" {
		role.Name = name
	}
	if code != "" {
		// Check uniqueness
		existing, _ := dao.GetRoleDAO().GetByCode(code)
		if existing != nil && existing.ID != id {
			return fmt.Errorf("role code '%s' already exists", code)
		}
		role.Code = code
	}
	if description != "" {
		role.Description = description
	}
	if status != nil {
		role.Status = *status
	}

	return dao.GetRoleDAO().Update(role)
}

func (s *roleServiceImpl) DeleteRole(id uint) error {
	return dao.GetRoleDAO().Delete(id)
}

func (s *roleServiceImpl) GetRole(id uint) (*model.Role, error) {
	return dao.GetRoleDAO().GetByID(id)
}

func (s *roleServiceImpl) ListRoles(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return dao.GetRoleDAO().List(page, pageSize, keyword)
}
