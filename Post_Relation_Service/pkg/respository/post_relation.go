package respository

import (
<<<<<<< HEAD
	"gorm.io/gorm"
	"github.com/Ansalps/Chattr_Post_Relation_service/pkg/repository/interfacesRepository"
=======
	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/repository/interfacesRepository"
	"gorm.io/gorm"
>>>>>>> develop
)

type PostRelationRepository struct {
	DB *gorm.DB
}

func NewPostRelationRepository(db *gorm.DB) interfacesRepository.PostRelationRepository {
	return &PostRelationRepository{
		DB: db,
	}
}
