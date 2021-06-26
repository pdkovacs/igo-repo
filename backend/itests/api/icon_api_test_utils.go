package api

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/pdkovacs/igo-repo/backend/pkg/domain"
	"github.com/pdkovacs/igo-repo/backend/pkg/web"
)

var backendSourceHome = os.Getenv("BACKEND_SOURCE_HOME")

func init() {
	if backendSourceHome == "" {
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.Getenv("USERPROFILE")
		}
		backendSourceHome = fmt.Sprintf("%s/github/pdkovacs/igo-repo/backend", homeDir)
	}
}

func getDemoIconfileContent(iconName string, iconfile domain.IconfileDescriptor) []byte {
	pathToContent := path.Join(backendSourceHome, "demo-data", iconfile.Format, iconfile.Size, fmt.Sprintf("%s.%s", iconName, iconfile.Format))
	content, err := os.ReadFile(pathToContent)
	if err != nil {
		panic(err)
	}
	return content
}

type testIconDescriptor struct {
	name       string
	modifiedBy string
	files      []domain.IconfileDescriptor
}

var testIconInputDataDescriptor = []testIconDescriptor{
	{
		name:       "attach_money",
		modifiedBy: "ux",
		files: []domain.IconfileDescriptor{
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
		name:       "cast_connected",
		modifiedBy: defaultCredentials.Username,
		files: []domain.IconfileDescriptor{
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

var moreTestIconInputDataDescriptor = []testIconDescriptor{
	{
		name:       "format_clear",
		modifiedBy: defaultCredentials.Username,
		files: []domain.IconfileDescriptor{
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
		name:       "insert_photo",
		modifiedBy: defaultCredentials.Username,
		files: []domain.IconfileDescriptor{
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

var dp2px = map[string]string{
	"24dp": "36px",
	"36dp": "54px",
}

var testIconDataResponse = []web.ResponseIcon{
	{
		Name:       "attach_money",
		ModifiedBy: getDefaultUserIDAsString(),
		Paths: []web.IconPath{
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "png", Size: "36px"}, Path: "/icon/attach_money/format/png/size/36px"},
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "svg", Size: "18px"}, Path: "/icon/attach_money/format/svg/size/18px"},
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "svg", Size: "24px"}, Path: "/icon/attach_money/format/svg/size/24px"},
		},
		Tags: []string{},
	},
	{
		Name:       "cast_connected",
		ModifiedBy: getDefaultUserIDAsString(),
		Paths: []web.IconPath{
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "png", Size: "36px"}, Path: "/icon/cast_connected/format/png/size/36px"},
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "svg", Size: "24px"}, Path: "/icon/cast_connected/format/svg/size/24px"},
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "svg", Size: "48px"}, Path: "/icon/cast_connected/format/svg/size/48px"},
		},
		Tags: []string{},
	},
}

var moreTestIconDataResponse = []web.ResponseIcon{
	{
		Name:       "format_clear",
		ModifiedBy: getDefaultUserIDAsString(),
		Paths: []web.IconPath{
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "png", Size: "36px"}, Path: "/icon/attach_money/format/png/size/36px"},
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "svg", Size: "48px"}, Path: "/icon/attach_money/format/svg/size/48px"},
		},
		Tags: []string{},
	},
	{
		Name:       "insert_photo",
		ModifiedBy: getDefaultUserIDAsString(),
		Paths: []web.IconPath{
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "png", Size: "36px"}, Path: "/icon/cast_connected/format/png/size/36px"},
			{IconfileDescriptor: domain.IconfileDescriptor{Format: "svg", Size: "48px"}, Path: "/icon/cast_connected/format/svg/size/48px"},
		},
		Tags: []string{},
	},
}

func createTestIconInputData(descriptors []testIconDescriptor) []domain.Icon {
	var icons = []domain.Icon{}

	for _, descriptor := range descriptors {

		var iconfiles = []domain.Iconfile{}

		for _, file := range descriptor.files {
			iconfile := createIconfile(file, getDemoIconfileContent(descriptor.name, file))
			iconfiles = append(iconfiles, iconfile)
		}

		icon := domain.Icon{
			IconAttributes: domain.IconAttributes{
				Name:       descriptor.name,
				ModifiedBy: descriptor.modifiedBy,
			},
			Iconfiles: iconfiles,
		}
		icons = append(icons, icon)
	}

	return icons
}

var testIconInputData []domain.Icon
var moreIconInputData []domain.Icon

func init() {
	testIconInputData = createTestIconInputData(testIconInputDataDescriptor)
	moreIconInputData = createTestIconInputData(moreTestIconInputDataDescriptor)
}

func createIconfile(desc domain.IconfileDescriptor, content []byte) domain.Iconfile {
	return domain.Iconfile{
		IconfileDescriptor: desc,
		Content:            content,
	}
}

type iconTestSuite struct {
	apiTestSuite
}

func (s *iconTestSuite) getCheckIconfile(session *apiTestSession, iconName string, iconfile domain.Iconfile) {
	actualIconfile, err := session.GetIconfile(iconName, iconfile.IconfileDescriptor)
	s.NoError(err)
	s.Equal(iconfile, actualIconfile)
}

func (s *iconTestSuite) assertGitCleanStatus() {
	s.testGitRepo.AssertGitCleanStatus(&s.Suite)
}

func (s *iconTestSuite) assertAllFilesInDBAreInGitAsWell() []string {
	checkedGitFiles := []string{}

	db := s.server.Repositories.DB
	git := s.testGitRepo

	allIconDesc, descAllErr := db.DescribeAllIcons()
	if descAllErr != nil {
		panic(descAllErr)
	}

	for _, iconDesc := range allIconDesc {
		for _, iconfileDesc := range iconDesc.Iconfiles {
			fileContentInDB, contentReadError := db.GetIconFile(iconDesc.Name, iconfileDesc.Format, iconfileDesc.Size)
			if contentReadError != nil {
				panic(contentReadError)
			}
			pathToFile := git.GetAbsolutePathToIconfile(iconDesc.Name, iconfileDesc)
			fileContentInGit, readGitFileErr := os.ReadFile(pathToFile)
			s.NoError(readGitFileErr)

			// TODO: fileContentInDB and fileContentInGit must equal
			s.True(bytes.Equal(fileContentInDB, fileContentInGit))

			checkedGitFiles = append(checkedGitFiles, s.testGitRepo.GetPathToIconfileInRepos(iconDesc.Name, iconfileDesc))
		}
	}

	return checkedGitFiles
}

func (s *iconTestSuite) assertAllFilesInGitAreInDBAsWell(iconfilesWithPeerInDB []string) {
	iconfiles, err := s.testGitRepo.GetIconfiles()
	s.NoError(err)
	for _, gitFile := range iconfiles {
		found := false
		for _, dbFile := range iconfilesWithPeerInDB {
			if gitFile == dbFile {
				found = true
				break
			}
		}
		if !found {
			s.Fail(fmt.Sprintf("%s doesn't have a peer in DB", gitFile))
		}
	}
}

func (s *iconTestSuite) assertReposInSync() {
	checkedGitFiles := s.assertAllFilesInDBAreInGitAsWell()
	s.assertAllFilesInGitAreInDBAsWell(checkedGitFiles)
}

func (s *iconTestSuite) assertEndState() {
	s.assertGitCleanStatus()
	s.assertReposInSync()
}
