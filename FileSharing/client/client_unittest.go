package client

///////////////////////////////////////////////////
//                                               //
// Everything in this file will NOT be graded!!! //
//                                               //
///////////////////////////////////////////////////

// In this unit tests file, you can write white-box unit tests on your implementation.
// These are different from the black-box integration tests in client_test.go,
// because in this unit tests file, you can use details specific to your implementation.

// For example, in this unit tests file, you can access struct fields and helper methods
// that you defined, but in the integration tests (client_test.go), you can only access
// the 8 functions (StoreFile, LoadFile, etc.) that are common to all implementations.

// In this unit tests file, you can write InitUser where you would write client.InitUser in the
// integration tests (client_test.go). In other words, the "client." in front is no longer needed.

import (
	"testing"

	userlib "github.com/cs161-staff/project2-userlib"

	_ "encoding/hex"
	//"encoding/json"

	_ "errors"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"

	_ "strconv"

	_ "strings"
)

func TestSetupAndExecution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Unit Tests")
}

var _ = Describe("Client Unit Tests", func() {

	BeforeEach(func() {
		userlib.DatastoreClear()
		userlib.KeystoreClear()
	})

	Describe("Unit Tests", func() {
		Specify("Basic Test: Check that the Username field is set for a new user", func() {
			userlib.DebugMsg("Initializing user Alice.")
			// Note: In the integration tests (client_test.go) this would need to
			// be client.InitUser, but here (client_unittests.go) you can write InitUser.
			alice, err := InitUser("alice", "password")
			Expect(err).To(BeNil())

			//check if a user doesn't exist
			_, err = GetUser("brodie", "nah")
			Expect(err).ToNot(BeNil())

			// Note: You can access the Username field of the User struct here.
			// But in the integration tests (client_test.go), you cannot access
			// struct fields because not all implementations will have a username field.
			Expect(alice.Username).To(Equal("alice"))
			_, err = GetUser("alice", "wrongpwd")

			Expect(err).ToNot(BeNil())

			newAlice, err := InitUser("alice", "password")
			Expect(newAlice).To(BeNil())

			bob, err := InitUser("bob", "password")
			Expect(err).To(BeNil())

			//test if users can share files among each other
			err = alice.StoreFile("Lebron.txt", []byte("king"))
			Expect(err).To(BeNil())
			err = bob.StoreFile("Lebron.tx", []byte("kin"))
			Expect(err).To(BeNil())

			idPointer, err := alice.CreateInvitation("Lebron.txt", "bob")
			Expect(err).To(BeNil())
			idPointer2, err := bob.CreateInvitation("Lebron.tx", "alice")
			Expect(err).To(BeNil())

			err = alice.AcceptInvitation("bob", idPointer2, "Lebron.tx")
			Expect(err).To(BeNil())
			err = bob.AcceptInvitation("alice", idPointer, "Lebron.txt")
			Expect(err).To(BeNil())

			giannis, err := InitUser("giannis", "passwprd")
			Expect(err).To(BeNil())
			idPointer, err = alice.CreateInvitation("Lebron.txt", "giannis")
			Expect(err).To(BeNil())
			idPointer2, err = bob.CreateInvitation("Lebron.tx", "giannis")
			Expect(err).To(BeNil())
			err = giannis.AcceptInvitation("bob", idPointer2, "Lebron.tx")

			fileBytes, err := alice.LoadFile("Lebron.tx")
			fileBytes, err = bob.LoadFile("Lebron.txt")
			fileBytes, err = giannis.LoadFile("Lebron.tx")
			Expect(fileBytes).To(Equal([]byte("kin")))

		})
	})

	Describe("Unit Tests for Files", func() {
		Specify("Basic Test: Check that the Username field is set for a new user", func() {
			userlib.DebugMsg("Initializing user Alice.")
			// Note: In the integration tests (client_test.go) this would need to
			// be client.InitUser, but here (client_unittests.go) you can write InitUser.
			alice, err := InitUser("alice", "password")
			Expect(err).To(BeNil())

			// Note: You can access the Username field of the User struct here.
			// But in the integration tests (client_test.go), you cannot access
			// struct fields because not all implementations will have a username field.
			Expect(alice.Username).To(Equal("alice"))
			_, err = GetUser("alice", "wrongpwd")

			Expect(err).ToNot(BeNil())

			err = alice.StoreFile("Lebron", []byte("king"))
			err = alice.StoreFile("Dame", []byte("king"))
			fileByte, err := alice.LoadFile("Lebron")
			Expect(fileByte).To(Equal([]byte("king")))

			//test overriding with store
			err = alice.StoreFile("Lebron", []byte("nevermind"))
			fileByte, err = alice.LoadFile("Lebron")
			Expect(fileByte).To(Equal([]byte("nevermind")))

			bob, err := InitUser("bob", "password")
			fileByte, err = bob.LoadFile("Lebron")
			Expect(err).ToNot(BeNil())

			//test append
			err = alice.AppendToFile("Lebron", []byte("me"))
			err = alice.AppendToFile("Lebron", []byte("yo"))
			fileByte, err = alice.LoadFile("Lebron")
			Expect(err).To(BeNil())

			err = bob.AppendToFile("Lebron", []byte("me"))
			fileByte, err = alice.LoadFile("Lebron")
			Expect(err).To(BeNil())

			//create and accept invitation normal
			idPointer, err := alice.CreateInvitation("Lebron", "bob")

			giannis, err := InitUser("giannis", "passwprd")
			giannisidPointer, err := alice.CreateInvitation("Lebron", "giannis")
			err = giannis.AcceptInvitation("alice", giannisidPointer, "Lebron")

			dameidPointer, err := alice.CreateInvitation("Dame", "bob")
			Expect(idPointer).ToNot(BeNil())

			err = bob.AcceptInvitation("alice", idPointer, "Lebron")
			fileByte, err = bob.LoadFile("Lebron")
			Expect(err).To(BeNil())

			cathy, err := InitUser("cathy", "password")
			idPointer, err = bob.CreateInvitation("Lebron", "cathy")

			err = cathy.AcceptInvitation("bob", idPointer, "Lebron")
			fileByte, err = cathy.LoadFile("Lebron")
			Expect(err).To(BeNil())

			//test revoke traditiona;
			//err = bob.RevokeAccess("Lebron", "cathy")
			david, err := InitUser("david", "passwprd")
			idPointer, err = david.CreateInvitation("Lebron", "cathy")

			fileByte, err = david.LoadFile("Lebron")
			Expect(err).ToNot(BeNil())

			err = alice.RevokeAccess("Lebron", "bob")
			fileByte, err = bob.LoadFile("Lebron")
			Expect(err).ToNot(BeNil())

			//make sure child can't acess file afer revoke
			fileByte, err = cathy.LoadFile("Lebron")
			Expect(err).ToNot(BeNil())

			idPointer, err = alice.CreateInvitation("Lebron", "david")
			//revoke before the other user accepts invitation
			err = alice.RevokeAccess("Lebron", "david")
			err = david.AcceptInvitation("bob", idPointer, "Lebron")
			fileByte, err = david.LoadFile("Lebron")

			err = bob.AcceptInvitation("alice", dameidPointer, "Dame")
			fileByte, err = bob.LoadFile("Dame")

			//we should get err for revoking a non existing or unshared file
			err = alice.RevokeAccess("Boi", "david")
			Expect(err).ToNot(BeNil())
			err = alice.RevokeAccess("Dame", "david")
			Expect(err).ToNot(BeNil())

			fileByte, err = alice.LoadFile("Lebron")
			err = alice.AppendToFile("Lebron", []byte("chill"))
			fileByte, err = alice.LoadFile("Lebron")
			Expect(fileByte).To(Equal([]byte("nevermindmeyochill")))

			fileByte, err = giannis.LoadFile("Lebron")
			Expect(err).To(BeNil())
			err = alice.AppendToFile("Lebron", []byte("chill"))
			Expect(err).To(BeNil())
			fileByte, err = giannis.LoadFile("Lebron")
			Expect(err).To(BeNil())
			Expect(fileByte).To(Equal([]byte("nevermindmeyochillchill")))

		})
	})

	Describe("Unit Tests for Files", func() {
		userlib.DebugMsg("Initializing user Alice.")
		// Note: In the integration tests (client_test.go) this would need to
		// be client.InitUser, but here (client_unittests.go) you can write InitUser.
		alice, err := InitUser("alice", "password")
		Expect(err).To(BeNil())

		// Note: You can access the Username field of the User struct here.
		// But in the integration tests (client_test.go), you cannot access
		// struct fields because not all implementations will have a username field.
		Expect(alice.Username).To(Equal("alice"))
		_, err = GetUser("alice", "wrongpwd")

		Expect(err).ToNot(BeNil())

		err = alice.StoreFile("Lebron.txt", []byte("king"))
		bob, err := InitUser("bob", "password")
		idPointer, err := alice.CreateInvitation("Lebron.txt", "bob")
		err = bob.AcceptInvitation("alice", idPointer, "Lebron.txt")
		fileByte, err := bob.LoadFile("Lebron.txt")

		Expect(err).To(BeNil())
		Expect(fileByte).To(Equal([]byte("king")))

		err = alice.RevokeAccess("Lebron.txt", "bob")
		Expect(err).To(BeNil())
		fileByte, err = bob.LoadFile("Lebron.txt")
		Expect(err).ToNot(BeNil())

		//fileBytes should be kingchill, but it's just showing king
		err = alice.AppendToFile("Lebron.txt", []byte("chill"))
		Expect(err).To(BeNil())
		err = alice.AppendToFile("Lebron.txt", []byte("chill"))
		Expect(err).To(BeNil())
		fileByte, err = alice.LoadFile("Lebron.txt")
		Expect(err).To(BeNil())
		Expect(fileByte).To(Equal([]byte("kingchillchill")))

	})
})
