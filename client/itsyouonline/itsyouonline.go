/*
 * Copyright (C) 2017-2018 GIG Technology NV and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package itsyouonline

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/itsyouonline/identityserver/clients/go/itsyouonline"
)

const (
	accessTokenURI = "https://itsyou.online/v1/oauth/access_token?response_type=id_token"
)

var (
	errNoPermission = errors.New("no permission")

	// ErrForbidden represents a forbidden action error
	ErrForbidden = errors.New("forbidden action")
)

// Config is used to create an IYO client.
type Config struct {
	Organization      string `yaml:"organization" json:"organization"`
	ApplicationID     string `yaml:"app_id" json:"app_id"`
	ApplicationSecret string `yaml:"app_secret" json:"app_secret"`
}

// Client defines itsyouonline client which is designed to help 0-stor user.
// It is not replacement for official itsyouonline client
type Client struct {
	cfg       Config
	iyoClient *itsyouonline.Itsyouonline
}

// NewClient creates new client
func NewClient(cfg Config) (*Client, error) {
	if cfg.Organization == "" {
		return nil, errors.New("IYO: organization not defined")
	}
	if cfg.ApplicationID == "" {
		return nil, errors.New("IYO: application ID not defined")
	}
	if cfg.ApplicationSecret == "" {
		return nil, errors.New("IYO: application Secret not defined")
	}
	return &Client{
		cfg:       cfg,
		iyoClient: itsyouonline.NewItsyouonline(),
	}, nil
}

// CreateJWT creates itsyouonline JWT token with these scopes:
// - org.namespace.read if perm.Read is true
// - org.namespace.write if perm.Write is true
// - org.namespace.delete if perm.Delete is true
func (c *Client) CreateJWT(namespace string, perm Permission) (string, error) {
	qp := map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     c.cfg.ApplicationID,
		"client_secret": c.cfg.ApplicationSecret,
		"validity":      "300", // 5 minutes, expressed in seconds
	}

	// build scopes query
	scopes := perm.Scopes(c.cfg.Organization, "0stor"+"."+namespace)
	if len(scopes) == 0 {
		return "", errNoPermission
	}
	qp["scope"] = strings.Join(scopes, ",")

	// create the request
	req, err := http.NewRequest("POST", accessTokenURI, nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = buildQueryString(req, qp)

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read response
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get access token, response code = %v", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	return string(b), err

}

// Creates name as suborganization of org
func createSubOrganization(c *Client, org, suborg string) *statusError {
	body := org + "." + suborg
	sub := itsyouonline.Organization{Globalid: body}

	_, resp, err := c.iyoClient.Organizations.CreateNewSubOrganization(org, sub, nil, nil)

	if err != nil {
		if resp.StatusCode == http.StatusConflict {
			err = fmt.Errorf("sub organization %s exists: %v", body, err)
		}
		return &statusError{
			Err:        err,
			StatusCode: resp.StatusCode,
		}
	}

	log.Infof("sub organization %s created", body)
	return nil
}

type statusError struct {
	Err        error
	StatusCode int
}

func (se *statusError) Error() string {
	return fmt.Sprintf("error response %d: %v", se.StatusCode, se.Err)
}

var (
	_ error = (*statusError)(nil)
)

// CreateNamespace creates namespace as itsyouonline organization
// Verifies the full namespace path exists, and creates it if don't
// It also creates....
// - org.0stor.namespace.read
// - org.0stor.namespace.write
// - org.0stor.namespace.write
func (c *Client) CreateNamespace(namespace string) error {
	err := c.login()
	if err != nil {
		return err
	}

	// Create c.cfg.Organization.0stor
	org := c.cfg.Organization
	statusErr := createSubOrganization(c, org, "0stor")
	if statusErr != nil && statusErr.StatusCode != http.StatusConflict {
		if statusErr.StatusCode != http.StatusForbidden {
			return err
		}

		// organization might not exist yet, let's verify that now
		_, resp, err := c.iyoClient.Organizations.GetOrganization(c.cfg.Organization, nil, nil)
		if err == nil {
			// something else must be wrong, simply return the statusErr
			return statusErr
		}
		if resp.StatusCode != http.StatusForbidden {
			return fmt.Errorf("GetOrganization failed code=%v, err=%v", resp.StatusCode, err)
		}

		organization := itsyouonline.Organization{Globalid: org}
		_, resp, err = c.iyoClient.Organizations.CreateNewOrganization(organization, nil, nil)
		if err != nil && resp.StatusCode != http.StatusConflict {
			return fmt.Errorf("failed to create new organization code=%v, err=%v", resp.StatusCode, err)
		}
		log.Infof("organization %s created", org)

		// Create c.cfg.Organization.0stor, once again,
		// if an error happens this times, we simply return it
		if statusErr = createSubOrganization(c, org, "0stor"); statusErr != nil {
			return statusErr
		}
	}

	// Create c.cfg.Organization.0stor.namespace
	org += ".0stor"
	if statusErr = createSubOrganization(c, org, namespace); statusErr != nil {
		return statusErr
	}

	// Create c.cfg.Organization.0stor.namespace. permissions
	org += "." + namespace
	permissions := Permission{Read: true, Write: true, Delete: true}
	for _, p := range permissions.perms() {
		if statusErr = createSubOrganization(c, org, p); statusErr != nil {
			return statusErr
		}
	}

	return nil
}

// DeleteNamespace deletes the namespace sub organization and all of it's sub organizations
func (c *Client) DeleteNamespace(namespace string) error {
	err := c.login()
	if err != nil {
		return err
	}

	resp, err := c.iyoClient.Organizations.DeleteOrganization(
		c.createNamespaceID(namespace), nil, nil)
	if err != nil {
		return fmt.Errorf(
			"deleting namespace failed: IYO returned status %+v \nwith error message: %v",
			resp.Status, err)
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	return nil
}

// GivePermission give a user some permission on a namespace
func (c *Client) GivePermission(namespace, userID string, perm Permission) error {
	err := c.login()
	if err != nil {
		return err
	}

	var org string
	for _, perm := range perm.perms() {
		if perm == "admin" {
			org = c.createNamespaceID(namespace)
		} else {
			org = c.createNamespaceID(namespace) + "." + perm
		}
		user := itsyouonline.OrganizationsGlobalidMembersPostReqBody{Searchstring: userID}
		_, resp, err := c.iyoClient.Organizations.AddOrganizationMember(org, user, nil, nil)
		if err != nil {
			return fmt.Errorf("give member permission failed: code=%v, err=%v", resp.StatusCode, err)
		}
		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("give member permission failed: code=%v", resp.StatusCode)
		}
	}

	return nil
}

// RemovePermission remove some permission from a user on a namespace
func (c *Client) RemovePermission(namespace, userID string, perm Permission) error {
	err := c.login()
	if err != nil {
		return err
	}

	var org string
	for _, perm := range perm.perms() {
		if perm == "admin" {
			org = c.createNamespaceID(namespace)
		} else {
			org = c.createNamespaceID(namespace) + "." + perm
		}
		resp, err := c.iyoClient.Organizations.RemoveOrganizationMember(userID, org, nil, nil)
		if err != nil {
			return fmt.Errorf("removing permission failed: IYO returned status %+v \nwith error message: %v", resp.Status, err)
		}
	}

	return nil
}

// GetPermission retrieves the permission a user has for a namespace
// returns true for a right when user is member of the namespace
func (c *Client) GetPermission(namespace, userID string) (*Permission, error) {
	err := c.login()
	if err != nil {
		return nil, err
	}

	var (
		permission = Permission{}
		org        string
	)

	allPermissions := []*struct {
		PropertyReference *bool
		String            string
		UseStringAsID     bool
	}{
		{&permission.Read, "read", true},
		{&permission.Write, "write", true},
		{&permission.Delete, "delete", true},
		{&permission.Admin, "admin", false},
	}

	for _, perm := range allPermissions {
		org = c.createNamespaceID(namespace)
		if perm.UseStringAsID {
			org += "." + perm.String
		}

		members, resp, err := c.iyoClient.Organizations.GetOrganizationUsers(org, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve user permission: %+v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Failed to retrieve user permission : IYO returned status %+v", resp.Status)
		}

		if !isMember(userID, members.Users) {
			continue
		}
		*perm.PropertyReference = true
	}
	return &permission, nil
}

func (c *Client) login() error {
	_, _, _, err := c.iyoClient.LoginWithClientCredentials(
		c.cfg.ApplicationID, c.cfg.ApplicationSecret)
	if err != nil {
		return fmt.Errorf("login failed:%v", err)
	}
	return nil
}

func (c *Client) createNamespaceID(namespace string) string {
	return c.cfg.Organization + "." + "0stor" + "." + namespace
}

func isMember(target string, list []itsyouonline.OrganizationUser) bool {
	for _, v := range list {
		if target == v.Username {
			return true
		}
	}
	return false
}

func buildQueryString(req *http.Request, qs map[string]interface{}) string {
	q := req.URL.Query()

	for k, v := range qs {
		q.Add(k, fmt.Sprintf("%v", v))
	}
	return q.Encode()
}
