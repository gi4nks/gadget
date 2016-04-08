package main

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jinzhu/gorm"
	"github.com/pivotal-golang/bytefmt"
)

type Repository struct {
	DB *gorm.DB
}

// HELPER FUNCTIONS
func repositoryFullName() string {
	return settings.RepositoryDirectory() + string(filepath.Separator) + settings.RepositoryFile()
}

//
func (r *Repository) InitDB() {
	var err error

	b, err := pathUtils.ExistsPath(settings.RepositoryDirectory())
	if err != nil {
		parrot.Error("Got error when reading repository directory", err)
	}

	if !b {
		pathUtils.CreatePath(settings.RepositoryDirectory())
	}

	r.DB, err = gorm.Open("sqlite3", repositoryFullName())
	if err != nil {
		parrot.Error("Got error creating repository directory", err)
	}
}

func (r *Repository) InitSchema() {

	if r.DB.HasTable(&ImageTag{}) {
		parrot.Debug("ImageTag already exists, removing it")
		r.DB.DropTable(&ImageTag{})
	}

	if r.DB.HasTable(&ImageLabel{}) {
		parrot.Debug("ImageLabel already exists, removing it")
		r.DB.DropTable(&ImageLabel{})
	}

	if r.DB.HasTable(&ImageBlob{}) {
		parrot.Debug("ImageBlob already exists, removing it")
		r.DB.DropTable(&ImageBlob{})
	}

	if r.DB.HasTable(&ImageVolume{}) {
		parrot.Debug("ImageVolume already exists, removing it")
		r.DB.DropTable(&ImageVolume{})
	}

	if r.DB.HasTable(&Image{}) {
		parrot.Debug("Image already exists, removing it")
		r.DB.DropTable(&Image{})
	}

	var il = &ImageLabel{}
	var it = &ImageTag{}
	var i = &Image{}
	var ib = &ImageBlob{}
	var iv = &ImageVolume{}

	r.DB.CreateTable(i)
	r.DB.CreateTable(il)
	r.DB.CreateTable(it)
	r.DB.CreateTable(ib)
	r.DB.CreateTable(iv)
}

func (r *Repository) CloseDB() {
	if err := r.DB.Close(); err != nil {
		parrot.Error("Error", err)
	}
}

func (r *Repository) BackupSchema() error {
	b, _ := pathUtils.ExistsPath(settings.RepositoryDirectory())
	if !b {
		return errors.New("Gadget repository path does not exist")
	}

	return pathUtils.CopyFile(repositoryFullName(), repositoryFullName()+".bkp")
}

// functionalities

func (r *Repository) Put(img docker.APIImages, imgDetails docker.Image) {
	parrot.Debug("[" + AsJson(img.RepoTags) + "] adding as " + TruncateID(img.ID))

	var image = Image{}
	image.ShortId = TruncateID(img.ID)
	image.LongId = img.ID

	image.CreatedAt = time.Unix(0, img.Created).Format("2006-01-02 15:04:05")
	image.Size = bytefmt.ByteSize(uint64(img.Size))
	image.VirtualSize = bytefmt.ByteSize(uint64(img.VirtualSize))

	image.Labels = []ImageLabel{}
	image.Tags = []ImageTag{}
	image.Volumes = []ImageVolume{}

	image.Blob = ImageBlob{}
	image.Blob.Summary = AsJson(img)

	// Adding tags
	for _, tag := range img.RepoTags {
		var imageTag = ImageTag{}

		imageTag.Name = strings.Split(tag, ":")[0]
		imageTag.Version = strings.Split(tag, ":")[1]
		imageTag.Tag = tag

		image.Tags = append(image.Tags, imageTag)
	}

	// Adding labels
	for k, v := range img.Labels {
		var imageLabel = ImageLabel{}

		imageLabel.Key = k
		imageLabel.Value = v
		imageLabel.Label = k + ":" + v

		image.Labels = append(image.Labels, imageLabel)
	}

	// Add volumes
	image.Blob.Details = AsJson(imgDetails)

	for k, v := range imgDetails.ContainerConfig.Volumes {
		var imageVolume = ImageVolume{}

		imageVolume.Volume = k
		imageVolume.Data = AsJson(v)

		image.Volumes = append(image.Volumes, imageVolume)
	}

	r.DB.Create(&image)
}

func (r *Repository) GetAll() []Image {
	images := []Image{}

	r.DB.Model(&images).Preload("Blob").Preload("Volumes").Preload("Tags").Preload("Labels").Find(&images)

	return images
}

func (r *Repository) Get(id string) Image {
	var image = Image{}

	r.DB.Model(&image).Where("short_id = ?", id).Preload("Blob").Preload("Volumes").Preload("Tags").Preload("Labels").First(&image)

	return image
}

func (r *Repository) Exists(id string) bool {
	var image = Image{}
	var count = -1

	r.DB.Where("short_id = ?", id).First(&image).Count(&count)

	if count == 0 {
		//parrot.Error("Error getting data", err)
		return false
	}
	parrot.Debug("Searching image with id", id, " - ", count)

	return true
}

func (r *Repository) FindByShortId(id string) Image {
	var image = Image{}

	r.DB.Model(&image).Where("short_id = ?", id).Preload("Blob").Preload("Volumes").Preload("Tags").Preload("Labels").First(&image)

	return image
}

func (r *Repository) FindByLongId(id string) Image {
	var image = Image{}

	r.DB.Model(&image).Where("long_id = ?", id).Preload("Blob").Preload("Volumes").Preload("Tags").Preload("Labels").First(&image)
	return image
}

func (r *Repository) FindByTag(tag string) Image {
	var image = Image{}
	var imageTag = ImageTag{}

	r.DB.Where("tag = ?", tag).First(&imageTag)

	if &imageTag == nil {
		parrot.Debug("No tag found")
		return image
	}

	r.DB.Model(&image).Where("id = ?", imageTag.ImageID).Preload("Volumes").Preload("Tags").Preload("Labels").Find(&image)

	return image
}

func (r *Repository) GetImagesWithLabels() []Image {
	images := []Image{}

	r.DB.Model(&images).Joins("inner join image_labels on image_labels.image_id = images.id").Preload("Volumes").Preload("Tags").Preload("Labels").Find(&images)

	return images
}

func (r *Repository) GetImagesByLabel(lbl string) []Image {
	images := []Image{}

	r.DB.Model(&images).Joins("inner join image_labels on image_labels.image_id = images.id").Where("label LIKE ?", "%"+lbl+"%").Preload("Tags").Preload("Labels").Find(&images)

	return images
}

func (r *Repository) GetImagesWithVolumes() []Image {
	images := []Image{}

	r.DB.Model(&images).Joins("inner join image_volumes on image_volumes.image_id = images.id").Preload("Volumes").Preload("Tags").Preload("Volumes").Find(&images)

	return images
}

func (r *Repository) GetImagesByVolume(vlm string) []Image {
	images := []Image{}

	r.DB.Model(&images).Joins("inner join image_volumes on image_volumes.image_id = images.id").Where("volume LIKE ?", "%"+vlm+"%").Preload("Tags").Preload("Volumes").Find(&images)

	return images
}
