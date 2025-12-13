package respository

import (
	"gorm.io/gorm"
	"github.com/Ansalps/Chattr_Post_Relation_service/pkg/repository/interfacesRepository"
)

type PostRelationRepository struct {
	DB *gorm.DB
}

func NewPostRelationRepository(db *gorm.DB) interfacesRepository.PostRelationRepository {
	return &PostRelationRepository{
		DB: db,
	}
}
