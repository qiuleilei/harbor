package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/replication/ng"
	"github.com/goharbor/harbor/src/replication/ng/dao"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

var (
	testRegistry = &model.Registry{
		Name: "test1",
		URL:  "https://test.harbor.io",
		Type: "harbor",
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "admin",
			AccessSecret: "Harbor12345",
		},
	}
	testRegistry2 = &model.Registry{
		Name: "test2",
		URL:  "https://test2.harbor.io",
		Type: "harbor",
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "admin",
			AccessSecret: "Harbor12345",
		},
	}
)

type RegistrySuite struct {
	suite.Suite
	testAPI         *testapi
	defaultRegistry model.Registry
}

func (suite *RegistrySuite) SetupSuite() {
	assert := assert.New(suite.T())
	assert.Nil(ng.Init(make(chan struct{})))

	suite.testAPI = newHarborAPI()
	code, err := suite.testAPI.RegistryCreate(*admin, testRegistry)
	assert.Nil(err)
	assert.Equal(http.StatusCreated, code)

	tmp, err := dao.GetRegistryByName(testRegistry.Name)
	assert.Nil(err)
	assert.NotNil(tmp)
	suite.defaultRegistry = *testRegistry
	suite.defaultRegistry.ID = tmp.ID

	CommonAddUser()
}

func (suite *RegistrySuite) TearDownSuite() {
	assert := assert.New(suite.T())
	code, err := suite.testAPI.RegistryDelete(*admin, suite.defaultRegistry.ID)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)

	CommonDelUser()
}

func (suite *RegistrySuite) TestGet() {
	assert := assert.New(suite.T())

	// Get a non-existed registry
	_, code, _ := suite.testAPI.RegistryGet(*admin, 0)
	assert.Equal(http.StatusBadRequest, code)

	// Get as admin, should succeed
	retrieved, code, err := suite.testAPI.RegistryGet(*admin, suite.defaultRegistry.ID)
	assert.Nil(err)
	assert.NotNil(retrieved)
	assert.Equal(http.StatusOK, code)
	assert.Equal("test1", retrieved.Name)

	// Get as user, should fail
	_, code, _ = suite.testAPI.RegistryGet(*testUser, suite.defaultRegistry.ID)
	assert.Equal(http.StatusForbidden, code)
}

func (suite *RegistrySuite) TestList() {
	assert := assert.New(suite.T())

	// List as admin, should succeed
	registries, code, err := suite.testAPI.RegistryList(*admin)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal(1, len(registries))

	// List as user, should fail
	registries, code, err = suite.testAPI.RegistryList(*testUser)
	assert.Equal(http.StatusForbidden, code)
	assert.Equal(0, len(registries))
}

func (suite *RegistrySuite) TestPost() {
	assert := assert.New(suite.T())

	// Should conflict when create exited registry
	code, err := suite.testAPI.RegistryCreate(*admin, testRegistry)
	assert.Nil(err)
	assert.Equal(http.StatusConflict, code)

	// Create as user, should fail
	code, err = suite.testAPI.RegistryCreate(*testUser, testRegistry2)
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, code)
}

func (suite *RegistrySuite) TestPing() {
	assert := assert.New(suite.T())

	code, err := suite.testAPI.RegistryPing(*admin, &model.Registry{
		ID: suite.defaultRegistry.ID,
	})
	assert.Nil(err)
	assert.Equal(http.StatusInternalServerError, code)

	code, err = suite.testAPI.RegistryPing(*admin, &model.Registry{
		ID: -1,
	})
	assert.Nil(err)
	assert.Equal(http.StatusNotFound, code)

	code, err = suite.testAPI.RegistryPing(*admin, &model.Registry{})
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, code)

	code, err = suite.testAPI.RegistryPing(*admin, &model.Registry{
		URL: "http://foo.bar",
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "admin",
			AccessSecret: "Harbor12345",
		},
	})
	assert.Nil(err)
	assert.NotEqual(http.StatusBadRequest, code)

	code, err = suite.testAPI.RegistryPing(*testUser, &model.Registry{
		ID: suite.defaultRegistry.ID,
	})
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, code)
}

func (suite *RegistrySuite) TestRegistryPut() {
	assert := assert.New(suite.T())

	// Update as admin, should succeed
	newKey := "NewKey"
	updateReq := &models.RegistryUpdateRequest{
		AccessKey: &newKey,
	}
	code, err := suite.testAPI.RegistryUpdate(*admin, suite.defaultRegistry.ID, updateReq)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	updated, code, err := suite.testAPI.RegistryGet(*admin, suite.defaultRegistry.ID)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("NewKey", updated.Credential.AccessKey)

	// Update as user, should fail
	code, err = suite.testAPI.RegistryUpdate(*testUser, suite.defaultRegistry.ID, updateReq)
	assert.NotNil(err)
	assert.Equal(http.StatusForbidden, code)
}

func (suite *RegistrySuite) TestDelete() {
	assert := assert.New(suite.T())

	code, err := suite.testAPI.RegistryCreate(*admin, testRegistry2)
	assert.Nil(err)
	assert.Equal(http.StatusCreated, code)

	tmp, err := dao.GetRegistryByName(testRegistry2.Name)
	assert.Nil(err)
	assert.NotNil(tmp)

	// Delete as user, should fail
	code, err = suite.testAPI.RegistryDelete(*testUser, tmp.ID)
	assert.NotNil(err)
	assert.Equal(http.StatusForbidden, code)

	// Delete as admin, should succeed
	code, err = suite.testAPI.RegistryDelete(*admin, tmp.ID)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
}

func TestRegistrySuite(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}