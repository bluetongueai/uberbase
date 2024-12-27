package deploy

import "strings"

type Environment struct {
	Variables map[string]string
	Files     []string
}

// Add a helper function to parse env_file
func parseEnvFile(envFile interface{}) []string {
	var result []string

	if envFile == nil {
		return result
	}

	switch v := envFile.(type) {
	case string:
		result = append(result, v)
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	case []string:
		return v
	}

	return result
}

// Add this helper function to parse environment from yaml.Node
func parseEnvironment(env interface{}) map[string]string {
	result := make(map[string]string)

	if env == nil {
		return result
	}

	switch v := env.(type) {
	case map[string]interface{}:
		for k, val := range v {
			if str, ok := val.(string); ok {
				result[k] = str
			}
		}
	case map[interface{}]interface{}:
		for k, val := range v {
			if ks, ok := k.(string); ok {
				if vs, ok := val.(string); ok {
					result[ks] = vs
				}
			}
		}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				parts := strings.SplitN(str, "=", 2)
				if len(parts) == 2 {
					result[parts[0]] = parts[1]
				}
			}
		}
	case map[string]string:
		return v
	case []string:
		for _, item := range v {
			parts := strings.SplitN(item, "=", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result
}
