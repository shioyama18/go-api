package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-api/recipes/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	redisClient *redis.Client
	ctx         context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		redisClient: redisClient,
		ctx:         ctx,
	}
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func (h *RecipesHandler) ListRecipeHandler(c *gin.Context) {
	if val, err := h.redisClient.Get(h.ctx, "recipes").Result(); err == redis.Nil {
		log.Printf("Cache missed")
		cur, err := h.collection.Find(h.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer cur.Close(h.ctx)
		recipes := make([]models.Recipe, 0)
		for cur.Next(h.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}
		data, _ := json.Marshal(recipes)
		h.redisClient.Set(h.ctx, "recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		log.Printf("Cache hit")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}

// swagger:operation POST /recipes recipes newRecipe
// Create a new recipe
// ---
// produces:
// - application/json
// responses:
//
//	'201':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func (h *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := h.collection.InsertOne(h.ctx, recipe)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting a new recipe",
		})
		return
	}
	data, _ := json.Marshal(recipe)
	h.redisClient.Set(h.ctx, recipe.ID.String(), string(data), 1*time.Hour)
	h.redisClient.Del(h.ctx, "recipes")
	c.JSON(http.StatusCreated, recipe.ID)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid recipe ID
func (h *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objectId}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: recipe.Name},
		{Key: "instructions", Value: recipe.Instructions},
		{Key: "ingredients", Value: recipe.Ingredients},
		{Key: "tags", Value: recipe.Tags},
	}}}
	opts := options.FindOneAndUpdate()
	err := h.collection.FindOneAndUpdate(h.ctx, filter, update, opts)

	if err.Err() != nil {
		if err.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Recipe not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Err().Error(),
			})
		}
		return
	}

	data, _ := json.Marshal(recipe)
	h.redisClient.Set(h.ctx, id, string(data), 1*time.Hour)
	h.redisClient.Del(h.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid recipe ID
func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	opts := options.FindOneAndDelete()
	err := h.collection.FindOneAndDelete(h.ctx, bson.M{"_id": objectId}, opts)

	if err.Err() != nil {
		if err.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Recipe not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Err().Error(),
			})
		}
		return
	}

	h.redisClient.Del(h.ctx, id)
	h.redisClient.Del(h.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

// swagger:operation GET /recipes/{id} recipes
// Get one recipe
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: recipe ID
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	    description: Successful operation
//	'404':
//	    description: Invalid recipe ID
func (h *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	if val, err := h.redisClient.Get(h.ctx, id).Result(); err == redis.Nil {
		log.Printf("Cache missed")
		objectId, _ := primitive.ObjectIDFromHex(id)
		cur := h.collection.FindOne(h.ctx, bson.M{
			"_id": objectId,
		})
		var recipe models.Recipe
		err := cur.Decode(&recipe)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		data, _ := json.Marshal(recipe)
		h.redisClient.Set(h.ctx, id, string(data), 0)
		c.JSON(http.StatusOK, recipe)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		log.Printf("Cache hit")
		var recipe models.Recipe
		json.Unmarshal([]byte(val), &recipe)
		c.JSON(http.StatusOK, recipe)
	}
}
