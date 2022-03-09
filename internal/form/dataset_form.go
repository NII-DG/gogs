package form

type DatasetFrom struct {
	Datasets []string `from:"dataset_list" binding:"Required"`
}

func (d *DatasetFrom) getDatasets() []string {
	return d.Datasets
}
