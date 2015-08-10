package s3resource

type Source struct {
	AccessKeyID         string `json:"access_key_id"`
	SecretAccessKey     string `json:"secret_access_key"`
	Bucket              string `json:"bucket"`
	Folder              string `json:"folder"`
	Filename            string `json:"filename"`
	Private             bool   `json:"private"`
	RegionName          string `json:"region_name"`
	CloudfrontURL       string `json:"cloudfront_url"`
	Endpoint            string `json:"endpoint"`
	DisableMD5HashCheck bool   `json:"disable_md5_hash_check"`

	VersionedFile string `json:"versioned_file"`
	Regexp        string `json:"regexp"`
}

func (source Source) IsValid() (bool, string) {
	if source.Folder == "" && source.Filename == "" {
		return false, "please specify either Folder or Filename"
	}

	return true, ""
}

type Version struct {
	Path      string `json:"path,omitempty"`
	VersionID string `json:"version_id,omitempty"`
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
