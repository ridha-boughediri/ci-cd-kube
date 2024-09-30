// internal/dto/object.go
package dto

import "encoding/xml"

type ListBucketResult struct {
	XMLName     xml.Name `xml:"ListBucketResult"`
	XMLNS       string   `xml:"xmlns,attr"`
	Name        string   `xml:"Name"`
	Prefix      string   `xml:"Prefix"`
	Marker      string   `xml:"Marker"`
	MaxKeys     int      `xml:"MaxKeys"`
	IsTruncated bool     `xml:"IsTruncated"`
	Contents    []Object `xml:"Contents"`
}

type Object struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}
