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
func jsonToParams(data map[string]interface{}, content *map[string]interface{}, featureID *int, tagIDs *[]int, isActive *bool) error {
	var ok bool
	*content, ok = data["content"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("error parsing content field")
	}
	float_val, ok := data["feature_id"].(float64)
	if !ok {
		return fmt.Errorf("error parsing feature_id field")
	}
	*featureID = int(float_val)
	ifaceSlice, ok := data["tag_ids"].([]interface{})
	if ok {
		*tagIDs = make([]int, len(ifaceSlice))
		for i, v := range ifaceSlice {
			(*tagIDs)[i] = int(v.(float64))
		}
	} else {
		return fmt.Errorf("error parsing tag_ids field")
	}
	*isActive, ok = data["is_active"].(bool)
	if !ok{
		return fmt.Errorf("error parsing is_active field")
	}
	return nil
}

// Wrapper function for building params for getBanner query
func getBannerQueryBuilder(params GetBannerParams) (string, []interface{}) {
	query := "SELECT * FROM banners WHERE 1=1"
	args := []interface{}{}
	count := 1
	if params.FeatureId != nil {
        query += fmt.Sprintf(" AND feature_id = $%d", count)
        args = append(args, *params.FeatureId)
		count++
    }
    if params.TagId != nil {
        query += fmt.Sprintf(" AND $%d = ANY(tag_ids)", count)
        args = append(args, *params.TagId)
		count++
    }
    if params.Limit != nil {
        query += fmt.Sprintf(" LIMIT $%d", count)
        args = append(args, *params.Limit)
		count++
    }
    if params.Offset != nil {
        query += fmt.Sprintf(" OFFSET $%d", count)
        args = append(args, *params.Offset)
		count++
    }
	return query, args
}
