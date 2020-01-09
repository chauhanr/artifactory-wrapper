package main

import "testing"

func TestPrepareUrl(t *testing.T) {
	repo := "mvn-local"
	groupId := "com.hcl.coe"
	artifactId := "qr"
	version := "0.0.1"
	fname := "jms-connect.jar"

	expectUrl := "http://localhost:8081/artifactory/mvn-local/com/hcl/coe/qr/0.0.1/jms-connect.jar"

	url := prepareArtifactoryUploadURL(repo, groupId, artifactId, version, fname)
	if url != expectUrl {
		t.Errorf("Expected value %s but got %s\n", expectUrl, url)
	}
}
