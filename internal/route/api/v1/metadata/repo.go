package metadata

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	datastruct "github.com/NII-DG/gogs/internal/route/api/v1/metadata/datastruct"
	"github.com/NII-DG/gogs/internal/urlutil"
	constval "github.com/NII-DG/gogs/internal/utils/const"
	"github.com/NII-DG/gogs/internal/utils/regex"
	log "unknwon.dev/clog/v2"
)

func GetAllMetadataByRepoIDAndBranch(c *context.APIContext) {
	repoid_str := c.Params(":repoid")
	if !regex.CheckNumeric(repoid_str) {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Repository ID is not Numeric.",
		})
		return
	}
	branch := c.Params(":branch")
	req_user := c.User
	log.Trace("API to get Research All Metadata[Repository ID : %s, Branch : %s] has been done by User[ID : %d]", repoid_str, branch, req_user.ID)

	repoid, _ := strconv.Atoi(repoid_str)
	repo, err := db.GetRepositoryByID(int64(repoid))
	if err != nil {
		err_msg := err.Error()
		if strings.Contains(err_msg, "repoID") {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": "Repository ID is not exist.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Internal Server Error",
		})
		log.Error("failure getting repository by owner name and repository name from DB.  Repository ID : %s", repoid_str)
		return
	}

	// check repository has branch
	if _, err := repo.GetBranch(branch); err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": fmt.Sprintf("this repository <ID : %s> dosen't have %s baranch.", repoid_str, branch),
		})
		return
	}

	// check request user access repository information
	users, err := repo.GetAssignees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Internal Server Error",
		})
		log.Error("failure getting read only user on repository from DB, Repository : %s", repo.Name)
		return
	}

	request_user_id := req_user.ID
	accessRight := false
	for _, u := range users {
		if u.ID == request_user_id {
			accessRight = true
		}
	}

	if !accessRight {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": fmt.Sprintf("you do not has access right to get repository <ID : %s> metadata.", repoid_str),
		})
		log.Trace("user<%s> do not has access right to get repository <ID : %s> metadata.", req_user.Name, repoid_str)
		return
	}

	// Create ResearchProject
	Research_pj := datastruct.ResearchProject{
		Name:        repo.ProtectName,
		Description: repo.ProjectDescription,
	}

	tmp_research_orgs := map[string]datastruct.ResearchOrg{}

	// Create Persons
	persons := []datastruct.Person{}
	for _, u := range users {

		personalUrl := ""
		if len(u.PersonalURL) > 0 {
			personalUrl = u.PersonalURL
		} else {
			url, err := urlutil.UpdatePath(c.BaseURL, u.Name)
			if err != nil {
				c.Errorf(err, "%v", err)
				return
			}
			personalUrl = url
		}

		aff_id := u.AffiliationId
		aff_db, err := db.GetAffiliationByID(aff_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"message": "Internal Server Error",
			})
			log.Error("failure getting affiliation from DB.  Affiliation ID : %d", aff_id)
			return

		}

		var research_org_id string
		if _, ok := tmp_research_orgs[aff_db.Url]; !ok {
			research_org := datastruct.ResearchOrg{
				ID:          aff_db.Url,
				Name:        aff_db.Name,
				Description: aff_db.Description,
				AliasName:   aff_db.Alias,
			}
			tmp_research_orgs[aff_db.Url] = research_org
			research_org_id = aff_db.Url
		} else {
			research_org_id = aff_db.Url
		}

		person := datastruct.Person{
			ID:                   u.IDStr(),
			Url:                  personalUrl,
			Name:                 u.FullName,
			Alias:                u.AliasName,
			Affiliation:          research_org_id,
			Email:                u.Email,
			Telephone:            u.Telephone,
			ERadResearcherNumber: u.ERadResearcherNumber,
		}
		persons = append(persons, person)
	}

	// Create Files and Dataset
	files, dataset, gin_monitoring, err := repo.ExtractMetadata(branch)
	if err != nil {
		log.Error("failure extracting metadata of data from repository <ID : %s>. err msg : %v", repoid_str, err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "Internal Server Error",
		})
		return
	}

	if gin_monitoring.DatasetStructure == "" {
		c.JSON(http.StatusNotFound, map[string]interface{}{
			"message": fmt.Sprintf("This repository <ID : %s, Name: %s> dosen't have Research Policy. Please 'ADD DMP' on Gin-fork UI", repoid_str, repo.FullName()),
		})
		return
	} else {
		if !constval.IsDatasetStructType(gin_monitoring.DatasetStructure) {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": fmt.Sprintf("Set Dataset Structure Type[%s] in this repository <ID : %s, Name: %s> is not invaid value. Please Edit DMP on Gin-fork UI", gin_monitoring.DatasetStructure, repoid_str, repo.FullName()),
			})
			return
		}
	}

	//TODO : Create dmps
	//TODO : Create funder_orgs
	//TODO : Create research_orgs
	//TODO : Create licenses
	//TODO : Create data_downloads
	//TODO : Create repository_objs

	//TODO : Create hosting_institutions

	// Create research_orgs
	research_orgs := []datastruct.ResearchOrg{}
	for _, v := range tmp_research_orgs {
		research_orgs = append(research_orgs, v)
	}

	repo_obj, err := repo.ExtractRepoMetadata()
	if err != nil {
		log.Error("failure extracting repository metadata from repository <ID : %s>. err msg : %v", repoid_str, err)
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"InternalServerError": "Internal Server Error",
		})
		return
	}

	// Create Metadata
	metadata := datastruct.Metadata{
		ResearchProject:     Research_pj,
		FunderOrgs:          []datastruct.FunderOrg{},
		ResearchOrgs:        research_orgs,
		Licenses:            []datastruct.License{},
		DataDownloads:       []datastruct.DataDownload{},
		RepositoryObjects:   []datastruct.RepositoryObject{repo_obj},
		HostingInstitutions: []datastruct.HostingInstitution{},
		Persons:             persons,
		Files:               files,
		Datasets:            dataset,
		GinMonitorings:      gin_monitoring,
		Dmps:                []datastruct.IFDmp{},
	}

	c.JSONSuccess(metadata)
}
