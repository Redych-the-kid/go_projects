package main

import "fmt"

// Checks if token is present in map
func validateToken(token string, tokens map[string]string) bool {
	if token != "" {
		return tokens[token] != ""
	} else {
		return false
	}
}

// Checks if token belongs to admin role
func validateAdminToken(token string, tokens map[string]string) bool {
	return tokens[token] == "admin"
}

// Parses JSON map and stores data in given parameters. Returns false if some parameteres weren't parsed sucessfully
func jsonToParams(data map[string]interface{}, content *map[string]interface{}, featureID *int, tagIDs *[]int, isActive *bool) bool {
	var ok bool
	*content, ok = data["content"].(map[string]interface{})
	if !ok {
		return false
	}
	float_val, ok := data["feature_id"].(float64)
	if !ok {
		return false
	}
	*featureID = int(float_val)
	ifaceSlice, ok := data["tag_ids"].([]interface{})
	if ok {
		*tagIDs = make([]int, len(ifaceSlice))
		for i, v := range ifaceSlice {
			(*tagIDs)[i] = int(v.(float64))
		}
	} else {
		return false
	}
	*isActive, ok = data["is_active"].(bool)
	return ok
}

// Wrapper function for building params for getBanner query
func getBannerQueryBuilder(params GetBannerParams) string {
	query := "SELECT * FROM banners"
	if params.FeatureId != nil || params.TagId != nil {
		query += " WHERE"
		before := false
		if params.FeatureId != nil {
			query += fmt.Sprintf(" feature_id = %d", *params.FeatureId)
			before = true
		}
		if params.TagId != nil {
			if before {
				query += " AND"
			}
			query += fmt.Sprintf(" %d = ANY(tag_ids)", *params.TagId)
		}
	}
	if params.Limit != nil {
		query += fmt.Sprintf(" LIMIT %d", *params.Limit)
	}
	if params.Offset != nil {
		query += fmt.Sprintf(" OFFSET %d", *params.Offset)
	}

	return query
}
