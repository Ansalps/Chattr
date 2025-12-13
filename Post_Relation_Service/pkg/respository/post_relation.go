package respository

import (
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/repository/interfacesRepository"
	"gorm.io/gorm"
)

type PostRelationRepository struct {
	DB *gorm.DB
}

func NewPostRelationRepository(db *gorm.DB) interfacesRepository.PostRelationRepository {
	return &PostRelationRepository{
		DB: db,
	}
}
