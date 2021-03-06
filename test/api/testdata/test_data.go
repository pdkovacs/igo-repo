package testdata

import (
	"fmt"
	"os"
	"path"

	"github.com/pdkovacs/igo-repo/api"
	"github.com/pdkovacs/igo-repo/config"
	"github.com/pdkovacs/igo-repo/domain"
	"github.com/pdkovacs/igo-repo/security/authn"
)

var backendSourceHome = os.Getenv("BACKEND_SOURCE_HOME")

var DefaultCredentials = config.PasswordCredentials{Username: "ux", Password: "ux"}
var defaultUserID = authn.LocalDomain.CreateUserID(DefaultCredentials.Username)

func init() {
	if backendSourceHome == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
		backendSourceHome = fmt.Sprintf("%s/github/pdkovacs/igo-repo", homeDir)
	}
}

func GetDemoIconfileContent(iconName string, iconfile domain.IconfileDescriptor) []byte {
	pathToContent := path.Join(backendSourceHome, "test/demo-data", iconfile.Format, iconfile.Size, fmt.Sprintf("%s.%s", iconName, iconfile.Format))
	content, err := os.ReadFile(pathToContent)
	if err != nil {
		panic(err)
	}
	return content
}

var testIconInputDataDescriptor = []domain.IconDescriptor{
	{
		IconAttributes: domain.IconAttributes{
			Name:       "attach_money",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
			{
				Format: "svg",
				Size:   "18px",
			},
			{
				Format: "svg",
				Size:   "24px",
			},
			{
				Format: "png",
				Size:   "24dp",
			},
		},
	},
	{
		IconAttributes: domain.IconAttributes{
			Name:       "cast_connected",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
			{
				Format: "svg",
				Size:   "24px",
			},
			{
				Format: "svg",
				Size:   "48px",
			},
			{
				Format: "png",
				Size:   "24dp",
			},
		},
	},
}

var moreTestIconInputDataDescriptor = []domain.IconDescriptor{
	{
		IconAttributes: domain.IconAttributes{
			Name:       "format_clear",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
			{
				Format: "png",
				Size:   "24dp",
			},
			{
				Format: "svg",
				Size:   "48px",
			},
		},
	},
	{
		IconAttributes: domain.IconAttributes{
			Name:       "insert_photo",
			ModifiedBy: defaultUserID.String(),
		},
		Iconfiles: []domain.IconfileDescriptor{
			{
				Format: "png",
				Size:   "24dp",
			},
			{
				Format: "svg",
				Size:   "48px",
			},
		},
	},
}

var DP2PX = map[string]string{
	"24dp": "36px",
	"36dp": "54px",
	"18px": "18px",
	"24px": "24px",
	"48px": "48px",
}

var testResponseIconData []api.ResponseIcon

var moreTestResponseIconData []api.ResponseIcon

func createTestIconInputData(descriptors []domain.IconDescriptor) []domain.Icon {
	var icons = []domain.Icon{}

	for _, descriptor := range descriptors {

		var iconfiles = []domain.Iconfile{}

		for _, file := range descriptor.Iconfiles {
			iconfile := createIconfile(file, GetDemoIconfileContent(descriptor.Name, file))
			iconfiles = append(iconfiles, iconfile)
		}

		icon := domain.Icon{
			IconAttributes: descriptor.IconAttributes,
			Iconfiles:      iconfiles,
		}
		icons = append(icons, icon)
	}

	return icons
}

var testIconInputDataMaster []domain.Icon
var moreTestIconInputDataMaster []domain.Icon

func mapIconfileSize(iconfile domain.IconfileDescriptor) domain.IconfileDescriptor {
	mappedIconfile := iconfile
	mappedSize, has := DP2PX[iconfile.Size]
	if !has {
		panic(fmt.Sprintf("Icon size %s cannot be mapped", iconfile.Size))
	}
	mappedIconfile.Size = mappedSize
	return mappedIconfile
}

func mapIconfileSizes(iconDescriptor domain.IconDescriptor) domain.IconDescriptor {
	newIconfiles := []domain.IconfileDescriptor{}
	for _, iconfile := range iconDescriptor.Iconfiles {
		newIconfiles = append(newIconfiles, mapIconfileSize(iconfile))
	}
	mappedIcon := iconDescriptor
	mappedIcon.Iconfiles = newIconfiles

	return mappedIcon
}

func init() {
	testIconInputDataMaster = createTestIconInputData(testIconInputDataDescriptor)
	moreTestIconInputDataMaster = createTestIconInputData(moreTestIconInputDataDescriptor)

	testResponseIconData = []api.ResponseIcon{}
	for _, testIconDescriptor := range testIconInputDataDescriptor {
		testResponseIconData = append(testResponseIconData, api.CreateResponseIcon("/icon", mapIconfileSizes(testIconDescriptor)))
	}

	moreTestResponseIconData = []api.ResponseIcon{}
	for _, testIconDescriptor := range moreTestIconInputDataDescriptor {
		moreTestResponseIconData = append(moreTestResponseIconData, api.CreateResponseIcon("/icon", mapIconfileSizes(testIconDescriptor)))
	}
}

func createIconfile(desc domain.IconfileDescriptor, content []byte) domain.Iconfile {
	return domain.Iconfile{
		IconfileDescriptor: desc,
		Content:            content,
	}
}

func getTestData(icons []domain.Icon, responseIcons []api.ResponseIcon) ([]domain.Icon, []api.ResponseIcon) {
	iconListClone := []domain.Icon{}
	for _, icon := range icons {
		iconfilesClone := make([]domain.Iconfile, len(icon.Iconfiles))
		tagsClone := make([]string, len(icon.Tags))
		copy(iconfilesClone, icon.Iconfiles)
		copy(tagsClone, icon.Tags)
		iconClone := domain.Icon{
			IconAttributes: domain.IconAttributes{
				Name:       icon.Name,
				ModifiedBy: icon.ModifiedBy,
				Tags:       tagsClone,
			},
			Iconfiles: iconfilesClone,
		}
		iconListClone = append(iconListClone, iconClone)
	}

	responseIconListClone := []api.ResponseIcon{}
	for _, resp := range responseIcons {
		paths := make([]api.IconPath, len(resp.Paths))
		tags := make([]string, len(resp.Tags))
		copy(paths, resp.Paths)
		copy(tags, resp.Tags)
		respClone := api.ResponseIcon{
			Name:       resp.Name,
			Paths:      paths,
			Tags:       tags,
			ModifiedBy: resp.ModifiedBy,
		}
		responseIconListClone = append(responseIconListClone, respClone)
	}

	return iconListClone, responseIconListClone
}

func Get() ([]domain.Icon, []api.ResponseIcon) {
	return getTestData(testIconInputDataMaster, testResponseIconData)
}

func GetMore() ([]domain.Icon, []api.ResponseIcon) {
	return getTestData(moreTestIconInputDataMaster, moreTestResponseIconData)
}
