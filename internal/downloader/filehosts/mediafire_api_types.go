package filehosts

/*
	https://www.mediafire.com/api/1.5/folder/get_content.php?content_type=CONTENT_TYPE&version=1.5&folder_key=FOLDER_KEY&response_format=json

	@param: FOLDER_KEY = string
	@param: CONTENT_TYPE = "files" | "folders"
*/

type MediafireFolderInfoResponse struct {
	Response struct {
		Action            string                 `json:"action"`
		Asynchronous      string                 `json:"asynchronous"`
		FolderContent     MediafireFolderContent `json:"folder_content"`
		Result            string                 `json:"result"`
		CurrentAPIVersion string                 `json:"current_api_version"`
	} `json:"response"`
}

type MediafireFileInfoResponse struct {
	Response struct {
		Action            string            `json:"action"`
		FileInfo          MediafireFileInfo `json:"file_info"`
		Result            string            `json:"result"`
		CurrentAPIVersion string            `json:"current_api_version"`
	} `json:"response"`
}

type MediafireFolderContent struct {
	ChunkSize   string                `json:"chunk_size"`
	ContentType string                `json:"content_type"`
	ChunkNumber string                `json:"chunk_number"`
	FolderKey   string                `json:"folderkey"`
	Files       []MediafireFileInfo   `json:"files,omitempty"`
	Folders     []MediafireFolderInfo `json:"folders,omitempty"`
	MoreChunks  string                `json:"more_chunks"`
	Revision    string                `json:"revision"`
}

type MediafireFileInfo struct {
	QuickKey          string               `json:"quickkey"`
	Ready             string               `json:"ready,omitempty"`
	Hash              string               `json:"hash"`
	Filename          string               `json:"filename"`
	Description       string               `json:"description"`
	Size              string               `json:"size"`
	Privacy           string               `json:"privacy"`
	Created           string               `json:"created"`
	PasswordProtected string               `json:"password_protected"`
	MimeType          string               `json:"mimetype"`
	FileType          string               `json:"filetype"`
	OwnerName         string               `json:"owner_name,omitempty"`
	View              string               `json:"view"`
	Edit              string               `json:"edit"`
	Revision          string               `json:"revision"`
	Flag              string               `json:"flag"`
	Permissions       MediafirePermissions `json:"permissions"`
	Downloads         string               `json:"downloads"`
	Views             string               `json:"views,omitempty"`
	Links             MediafireLinks       `json:"links"`
	CreatedUTC        string               `json:"created_utc"`
}

type MediafireFolderInfo struct {
	FolderKey      string               `json:"folderkey"`
	Name           string               `json:"name"`
	Description    string               `json:"description"`
	Tags           string               `json:"tags"`
	Privacy        string               `json:"privacy"`
	Created        string               `json:"created"`
	Revision       string               `json:"revision"`
	Flag           string               `json:"flag"`
	Permissions    MediafirePermissions `json:"permissions"`
	FileCount      string               `json:"file_count"`
	FolderCount    string               `json:"folder_count"`
	DropboxEnabled string               `json:"dropbox_enabled"`
	CreatedUTC     string               `json:"created_utc"`
}

type MediafirePermissions struct {
	Value    string `json:"value"`
	Explicit string `json:"explicit"`
	Read     string `json:"read"`
	Write    string `json:"write"`
}

type MediafireLinks struct {
	View           string `json:"view,omitempty"`
	NormalDownload string `json:"normal_download"`
}
