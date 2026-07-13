// Copyright (c) 2026 Guillaume Evrard
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:generate go tool oapi-codegen -config oapi-config.yaml openapi.yaml

package api

import (
	"context"
	"jij-service/domain"
)

type Server struct {
	ProfileService domain.ProfileService
}

func (s *Server) GetLeaderboardLeaderboardName(_ context.Context, requestObject GetLeaderboardLeaderboardNameRequestObject) (GetLeaderboardLeaderboardNameResponseObject, error) {
	response := GetLeaderboardLeaderboardName200JSONResponse{
		Entries: new(make([]LeaderboardEntry, 10)),
	}
	return response, nil
}

func (s *Server) GetPing(_ context.Context, _ GetPingRequestObject) (GetPingResponseObject, error) {
	response := GetPing200JSONResponse{}
	return response, nil
}

func (s *Server) GetProfileProfileIdComponentComponentId(ctx context.Context, requestObject GetProfileProfileIdComponentComponentIdRequestObject) (GetProfileProfileIdComponentComponentIdResponseObject, error) {
	componentData, err := s.ProfileService.GetProfileComponent(ctx, requestObject.ProfileId, requestObject.ComponentId)
	response := GetProfileProfileIdComponentComponentId200JSONResponse{
		ComponentId: new(requestObject.ComponentId),
		Data:        new(ProfileComponentData(componentData)),
	}
	return response, err
}

func (s *Server) PatchProfileProfileIdComponentComponentId(ctx context.Context, requestObject PatchProfileProfileIdComponentComponentIdRequestObject) (PatchProfileProfileIdComponentComponentIdResponseObject, error) {
	componentData, err := s.ProfileService.PatchProfileComponent(ctx, requestObject.ProfileId, requestObject.ComponentId, domain.ProfileComponentDataDelta(*requestObject.Body.Data))
	response := PatchProfileProfileIdComponentComponentId200JSONResponse{
		ComponentId: new(requestObject.ComponentId),
		Data:        new(ProfileComponentData(componentData)),
	}
	return response, err
}

func (s *Server) PostProfileProfileIdRunRunId(ctx context.Context, _ PostProfileProfileIdRunRunIdRequestObject) (PostProfileProfileIdRunRunIdResponseObject, error) {
	response := PostProfileProfileIdRunRunId200Response{}
	return response, nil
}

// quests/missions etc
// notifications
