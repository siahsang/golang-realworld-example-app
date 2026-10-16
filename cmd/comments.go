package main

//
//func prepareSingleCommentsResponse(app *application, r *http.Request, comment *models.Comment, loginUser *auth.User) (envelope, error) {
//	return prepareCommentsResponse(app, r, []*models.Comment{comment}, loginUser, true)
//}
//
//func prepareMultiCommentsResponse(app *application, r *http.Request, comments []*models.Comment, loginUser *auth.User) (envelope, error) {
//	return prepareCommentsResponse(app, r, comments, loginUser, false)
//}
//
//func prepareCommentsResponse(app *application, r *http.Request, comments []*models.Comment, loginUser *auth.User, singComment bool) (envelope, error) {
//	var response []CommentResponse
//
//	for _, comment := range comments {
//		commentResponse := CommentResponse{}
//		profile, err := app.core.GetProfileByUserId(r.Context(), comment.AuthorID)
//		if err != nil {
//			return nil, err
//		}
//
//		commentResponse.ID = comment.ID
//		commentResponse.Body = comment.Body
//		commentResponse.CreatedAt = comment.CreatedAt
//		commentResponse.UpdatedAt = comment.UpdatedAt
//
//		if loginUser != nil {
//			commentResponse.Author = &CommentAuthorBody{}
//			// Map the fields from models.Comment to CommentResponse
//			commentResponse.Author.Bio = profile.Bio
//			commentResponse.Author.Following = profile.Following
//			commentResponse.Author.Image = profile.Image
//			commentResponse.Author.Username = profile.Username
//		} else {
//			commentResponse.Author = nil
//		}
//		response = append(response, commentResponse)
//	}
//
//	if singComment {
//		return envelope{"comment": response[0]}, nil
//	} else {
//		return envelope{"comments": response}, nil
//	}
//}
