package uid

import "testing"

func Test_New(t *testing.T) {

	testCases := []struct {
		UserID     UserID
		SourceType SourceType
		SourceID   uint64
		Expected   string
	}{
		{1, 1, 1, "01011"},
		{2, Twitter, 1604043506523295746, "0200c6q1yu1rxgci"},
		{23, Twitter, 33234523452345433, "0n00938nlatk57d"},
	}

	for _, testCase := range testCases {
		actual := New(testCase.UserID, testCase.SourceType, testCase.SourceID)
		if actual.String() != testCase.Expected {
			t.Errorf("Expected %s, Actual %s", testCase.Expected, actual)
		}
	}
}
