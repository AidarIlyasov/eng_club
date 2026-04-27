package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/valyala/fasthttp"

	"eng_club/models"
)

// ListPlaces handles GET /api/members.
func (d *Deps) ListMembers(ctx *fasthttp.RequestCtx) {
	members, err := d.DB.ListMembers()
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if members == nil {
		members = []models.Member{}
	}
	WriteJSON(ctx, fasthttp.StatusOK, members)
}

func (d *Deps) ListWithoutPairs(ctx *fasthttp.RequestCtx) {
	members, err := d.DB.ListWithoutPairs()
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	if members == nil {
		members = []models.Member{}
	}
	WriteJSON(ctx, fasthttp.StatusOK, members)
}

func (d *Deps) AddWithoutPairs(ctx *fasthttp.RequestCtx) {
	var memberID int
	var err error
	var body struct {
		ID       string `json:"id,omitempty"`
		Name     string `json:"name"`
		Telegram string `json:"telegram"`
	}

	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		WriteErr(ctx, fasthttp.StatusBadRequest, "invalid JSON body")
		return
	}

	if body.Name == "" || body.Telegram == "" {
		WriteErr(ctx, fasthttp.StatusBadRequest, "name and telegram are required")
		return
	}

	// Upsert member
	if body.ID == "" {
		memberID, err = d.DB.UpsertMember(body.Name, body.Telegram, "", nil)
		if err != nil {
			WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
	} else {
		memberID, err = strconv.Atoi(body.ID)
		if err != nil {
			WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
			return
		}
	}

	err = d.DB.AddWithoutPairs(memberID)
	if err != nil {
		WriteErr(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}

	WriteJSON(ctx, fasthttp.StatusOK, map[string]any{
		"success": true,
	})
}
