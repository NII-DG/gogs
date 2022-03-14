package form

type DatasetFrom struct {
	DatasetList []string `binding:"Required"`
}

func (d *DatasetFrom) getDatasets() []string {
	return d.DatasetList
}
