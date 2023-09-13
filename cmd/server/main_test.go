package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shioyama18/go-api/models"
	"github.com/stretchr/testify/assert"
)

func TestListRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(SetupServer(false))
	defer ts.Close()
	resp, err := http.Get(fmt.Sprintf("%s/recipes", ts.URL))
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	data, _ := io.ReadAll(resp.Body)
	var recipes []models.Recipe
	json.Unmarshal(data, &recipes)
	assert.Equal(t, 492, len(recipes))
}

func TestFindRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(SetupServer(false))
	defer ts.Close()

	expectedRecipe := models.Recipe{
		Name: "Oregano Marinated Chicken",
		Tags: []string{"main", "chicken"},
	}
	id := "6501b32bad3488e6f1c6aa5f"

	resp, err := http.Get(fmt.Sprintf("%s/recipes/%s", ts.URL, id))
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	data, _ := io.ReadAll(resp.Body)
	var actualRecipe models.Recipe
	json.Unmarshal(data, &actualRecipe)
	assert.Equal(t, expectedRecipe.Name, actualRecipe.Name)
	assert.Equal(t, len(expectedRecipe.Tags), len(actualRecipe.Tags))
}

func TestUpdateRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(SetupServer(false))
	defer ts.Close()
	id := "6501b32bad3488e6f1c6aa60"
	recipe := models.Recipe{
		Name: "Green pea soup without cheddar scallion panini",
	}
	raw, _ := json.Marshal(recipe)
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/recipes/%s", ts.URL, id),
		bytes.NewBuffer(raw),
	)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	data, _ := io.ReadAll(resp.Body)
	var payload map[string]string
	json.Unmarshal(data, &payload)
	assert.Equal(t, payload["message"], "Recipe has been updated")
}

func TestDeleteRecipeHandler(t *testing.T) {
	ts := httptest.NewServer(SetupServer(false))
	defer ts.Close()
	id := "6501b32bad3488e6f1c6aa61"
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/recipes/%s", ts.URL, id),
		nil,
	)
	assert.Nil(t, err)
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	data, _ := io.ReadAll(resp.Body)
	var payload map[string]string
	json.Unmarshal(data, &payload)
	assert.Equal(t, payload["message"], "Recipe has been deleted")
}
