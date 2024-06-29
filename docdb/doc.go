package vclouddb

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EulaStatus int

const (
	Pending    EulaStatus = iota // 0
	Uploaded                     // 1
	Available                    // 2
	Deprecated                   // 3
)

type EulaType int

const (
	ContentCreator EulaStatus = iota // 0
	EndCustomer                      // 1
	PrivacyPolicy                    // 2
)
package vclouddb

import (
"fmt"
"github.com/google/uuid"
"gorm.io/gorm"
)

type EulaStatus int

const (
	Pending    EulaStatus = iota // 0
	Uploaded                     // 1
	Available                    // 2
	Deprecated                   // 3
)

type EulaType int

const (
	ContentCreator EulaStatus = iota // 0
	EndCustomer                      // 1
	PrivacyPolicy                    // 2
)

type Eula struct {
	Model
	Version        int
	Content        string
	FilePath       string
	Status         EulaStatus // Enum to indicate the current status
	UploadUserUUID uuid.UUID  // one-to-one relationship
	Type           EulaType
}

func (e *Eula) String() string {
	return fmt.Sprintf("Eula<%+v>", *e)
}

func (e *Eula) BeforeCreate(tx *gorm.DB) (err error) {
	e.UUID = uuid.New()
	return
}

func (vdb *VCloudDb) CreateEula(eula Eula) (*Eula, error) {
	if err := vdb.db.Create(&eula).Error; err != nil {
		return nil, err
	}
	return &eula, nil
}

func (vdb *VCloudDb) FindEulaByVersion(version int) (*Eula, error) {
	var eula Eula
	if err := vdb.db.Where("version = ?", version).First(&eula).Error; err != nil {
		return nil, err
	}
	return &eula, nil
}

func (vdb *VCloudDb) UpdateEulaStatus(version int, status EulaStatus) (*Eula, error) {
	eula, err := vdb.FindEulaByVersion(version)
	if err != nil {
		return nil, err
	}
	eula.Status = status
	if err := vdb.db.Save(&eula).Error; err != nil {
		return nil, err
	}
	return eula, nil
}

type Eula struct {
	Model
	Version        int
	Content        string
	FilePath       string
	Status         EulaStatus // Enum to indicate the current status
	UploadUserUUID uuid.UUID  // one-to-one relationship
	Type           EulaType
}

func (e *Eula) String() string {
	return fmt.Sprintf("Eula<%+v>", *e)
}

func (e *Eula) BeforeCreate(tx *gorm.DB) (err error) {
	e.UUID = uuid.New()
	return
}

func (vdb *VCloudDb) CreateEula(eula Eula) (*Eula, error) {
	if err := vdb.db.Create(&eula).Error; err != nil {
		return nil, err
	}
	return &eula, nil
}

func (vdb *VCloudDb) FindEulaByVersion(version int) (*Eula, error) {
	var eula Eula
	if err := vdb.db.Where("version = ?", version).First(&eula).Error; err != nil {
		return nil, err
	}
	return &eula, nil
}

func (vdb *VCloudDb) UpdateEulaStatus(version int, status EulaStatus) (*Eula, error) {
	eula, err := vdb.FindEulaByVersion(version)
	if err != nil {
		return nil, err
	}
	eula.Status = status
	if err := vdb.db.Save(&eula).Error; err != nil {
		return nil, err
	}
	return eula, nil
}
