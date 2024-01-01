package client_test

// You MUST NOT change these default imports.  ANY additional imports may
// break the autograder and everyone will be sad.

import (
	// Some imports use an underscore to prevent the compiler from complaining
	// about unused imports.
	_ "encoding/hex"
	_ "errors"
	_ "strconv"
	_ "strings"
	"testing"

	// A "dot" import is used here so that the functions in the ginko and gomega
	// modules can be used without an identifier. For example, Describe() and
	// Expect() instead of ginko.Describe() and gomega.Expect().
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	userlib "github.com/cs161-staff/project2-userlib"

	"github.com/cs161-staff/project2-starter-code/client"
)

func TestSetupAndExecution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Tests")
}

// ================================================
// Global Variables (feel free to add more!)
// ================================================
const defaultPassword = "password"
const emptyString = ""
const contentOne = "Bitcoin is Nick's favorite "
const contentTwo = "digital "
const contentThree = "cryptocurrency!"

// ================================================
// Describe(...) blocks help you organize your tests
// into functional categories. They can be nested into
// a tree-like structure.
// ================================================

var _ = Describe("Client Tests", func() {

	// A few user declarations that may be used for testing. Remember to initialize these before you
	// attempt to use them!
	var alice *client.User
	var bob *client.User
	var charles *client.User
	// var doris *client.User
	// var eve *client.User
	// var frank *client.User
	// var grace *client.User
	// var horace *client.User
	// var ira *client.User

	// These declarations may be useful for multi-session testing.
	var alicePhone *client.User
	var aliceLaptop *client.User
	var aliceDesktop *client.User

	var err error

	// A bunch of filenames that may be useful.
	aliceFile := "aliceFile.txt"
	bobFile := "bobFile.txt"
	charlesFile := "charlesFile.txt"
	// dorisFile := "dorisFile.txt"
	// eveFile := "eveFile.txt"
	// frankFile := "frankFile.txt"
	// graceFile := "graceFile.txt"
	// horaceFile := "horaceFile.txt"
	// iraFile := "iraFile.txt"

	BeforeEach(func() {
		// This runs before each test within this Describe block (including nested tests).
		// Here, we reset the state of Datastore and Keystore so that tests do not interfere with each other.
		// We also initialize
		userlib.DatastoreClear()
		userlib.KeystoreClear()
	})

	//Attacker: Revoke user, when that user tries to read that file should be error
	//fuzz-testing: random writes to random locations,  tamper with random locations in datastore, cover cases we don't think of ex:
	//testing swapping entries in datastore, the attacker can swap entries in datastore - make sure the encryption and decryption verfies the correct key of datastore
	//getUser: another user logs in with the wrong password, should be rror

	Describe("Basic Tests", func() {

		Specify("Basic Test: Testing InitUser/GetUser on a single user.", func() {
			userlib.DebugMsg("Initializing user Alice.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Getting user Alice.")
			aliceLaptop, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())
		})

		Specify("Basic Test: Testing Single User Store/Load/Append.", func() {
			userlib.DebugMsg("Initializing user Alice.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Storing file data: %s", contentOne)
			err = alice.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Appending file data: %s", contentTwo)
			err = alice.AppendToFile(aliceFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Appending file data: %s", contentThree)
			err = alice.AppendToFile(aliceFile, []byte(contentThree))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Loading file...")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))
		})

		Specify("Basic Test: Testing Create/Accept Invite Functionality with multiple users and multiple instances.", func() {
			userlib.DebugMsg("Initializing users Alice (aliceDesktop) and Bob.")
			aliceDesktop, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Getting second instance of Alice - aliceLaptop")
			aliceLaptop, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceDesktop storing file %s with content: %s", aliceFile, contentOne)
			err = aliceDesktop.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceLaptop creating invite for Bob.")
			invite, err := aliceLaptop.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob accepting invite from Alice under filename %s.", bobFile)
			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob appending to file %s, content: %s", bobFile, contentTwo)
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceDesktop appending to file %s, content: %s", aliceFile, contentThree)
			err = aliceDesktop.AppendToFile(aliceFile, []byte(contentThree))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that aliceDesktop sees expected file data.")
			data, err := aliceDesktop.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Checking that aliceLaptop sees expected file data.")
			data, err = aliceLaptop.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Checking that Bob sees expected file data.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Getting third instance of Alice - alicePhone.")
			alicePhone, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that alicePhone sees Alice's changes.")
			data, err = alicePhone.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))
		})

		Specify("Basic Test: Testing Revoke Functionality", func() {
			userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			charles, err = client.InitUser("charles", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
			alice.StoreFile(aliceFile, []byte(contentOne))

			userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)

			invite, err := alice.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
			invite, err = bob.CreateInvitation(bobFile, "charles")
			Expect(err).To(BeNil())

			err = charles.AcceptInvitation("bob", invite, charlesFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Charles can load the file.")
			data, err = charles.LoadFile(charlesFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Alice revoking Bob's access from %s.", aliceFile)
			err = alice.RevokeAccess(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err = alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob/Charles lost access to the file.")
			_, err = bob.LoadFile(bobFile)
			Expect(err).ToNot(BeNil())

			_, err = charles.LoadFile(charlesFile)
			Expect(err).ToNot(BeNil())

			userlib.DebugMsg("Checking that the revoked users cannot append to the file.")
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())

			err = charles.AppendToFile(charlesFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())

		})

		Specify("Basic Test: Testing Revoke Functionality", func() {
			userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			charles, err = client.InitUser("charles", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
			alice.StoreFile(aliceFile, []byte(contentOne))

			userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)

			invite, err := alice.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
			invite, err = bob.CreateInvitation(bobFile, "charles")
			Expect(err).To(BeNil())

			err = charles.AcceptInvitation("bob", invite, charlesFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Charles can load the file.")
			data, err = charles.LoadFile(charlesFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Alice revoking Bob's access from %s.", aliceFile)
			err = alice.RevokeAccess(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err = alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob/Charles lost access to the file.")
			_, err = bob.LoadFile(bobFile)
			Expect(err).ToNot(BeNil())

			_, err = charles.LoadFile(charlesFile)
			Expect(err).ToNot(BeNil())

			userlib.DebugMsg("Checking that the revoked users cannot append to the file.")
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())

			err = charles.AppendToFile(charlesFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())

		})

		Specify("Test multiple shares", func() {
			userlib.DebugMsg("Initializing user Alice.")
			// Note: In the integration tests (client_test.go) this would need to
			// be client.InitUser, but here (client_unittests.go) you can write InitUser.
			alice, err := client.InitUser("alice", "password")
			Expect(err).To(BeNil())

			//check if a user doesn't exist
			_, err = client.GetUser("brodie", "nah")
			Expect(err).ToNot(BeNil())

			// Note: You can access the Username field of the User struct here.
			// But in the integration tests (client_test.go), you cannot access
			// struct fields because not all implementations will have a username field.
			Expect(alice.Username).To(Equal("alice"))
			_, err = client.GetUser("alice", "wrongpwd")

			Expect(err).ToNot(BeNil())

			newAlice, err := client.InitUser("alice", "password")
			Expect(newAlice).To(BeNil())

			bob, err := client.InitUser("bob", "password")
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

			giannis, err := client.InitUser("giannis", "passwprd")
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

		Specify("Test Append and File Sharing", func() {
			userlib.DebugMsg("Initializing user Alice.")
			// Note: In the integration tests (client_test.go) this would need to
			// be client.InitUser, but here (client_unittests.go) you can write InitUser.
			alice, err := client.InitUser("alice", "password")
			Expect(err).To(BeNil())

			// Note: You can access the Username field of the User struct here.
			// But in the integration tests (client_test.go), you cannot access
			// struct fields because not all implementations will have a username field.
			Expect(alice.Username).To(Equal("alice"))
			_, err = client.GetUser("alice", "wrongpwd")

			Expect(err).ToNot(BeNil())

			err = alice.StoreFile("Lebron", []byte("king"))
			err = alice.StoreFile("Dame", []byte("king"))
			fileByte, err := alice.LoadFile("Lebron")
			Expect(fileByte).To(Equal([]byte("king")))

			//test overriding with store
			err = alice.StoreFile("Lebron", []byte("nevermind"))
			fileByte, err = alice.LoadFile("Lebron")
			Expect(fileByte).To(Equal([]byte("nevermind")))

			bob, err := client.InitUser("bob", "password")
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

			giannis, err := client.InitUser("giannis", "passwprd")
			giannisidPointer, err := alice.CreateInvitation("Lebron", "giannis")
			err = giannis.AcceptInvitation("alice", giannisidPointer, "Lebron")

			dameidPointer, err := alice.CreateInvitation("Dame", "bob")
			Expect(idPointer).ToNot(BeNil())

			err = bob.AcceptInvitation("alice", idPointer, "Lebron")
			fileByte, err = bob.LoadFile("Lebron")
			Expect(err).To(BeNil())

			cathy, err := client.InitUser("cathy", "password")
			idPointer, err = bob.CreateInvitation("Lebron", "cathy")

			err = cathy.AcceptInvitation("bob", idPointer, "Lebron")
			fileByte, err = cathy.LoadFile("Lebron")
			Expect(err).To(BeNil())

			//test revoke traditiona;
			//err = bob.RevokeAccess("Lebron", "cathy")
			david, err := client.InitUser("david", "passwprd")
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

		Specify("Unit Tests for Files", func() {
			userlib.DebugMsg("Initializing user Alice.")
			// Note: In the integration tests (client_test.go) this would need to
			// be client.InitUser, but here (client_unittests.go) you can write InitUser.
			alice, err := client.InitUser("alice", "password")
			Expect(err).To(BeNil())

			// Note: You can access the Username field of the User struct here.
			// But in the integration tests (client_test.go), you cannot access
			// struct fields because not all implementations will have a username field.
			Expect(alice.Username).To(Equal("alice"))
			_, err = client.GetUser("alice", "wrongpwd")

			Expect(err).ToNot(BeNil())

			err = alice.StoreFile("Lebron.txt", []byte("king"))
			bob, err := client.InitUser("bob", "password")
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
})
