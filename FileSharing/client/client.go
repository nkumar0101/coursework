package client

// CS 161 Project 2

// Only the following imports are allowed! ANY additional imports
// may break the autograder!
// - bytes
// - encoding/hex
// - encoding/json
// - errors
// - fmt
// - github.com/cs161-staff/project2-userlib
// - github.com/google/uuid
// - strconv
// - strings

import (
	"encoding/json"

	userlib "github.com/cs161-staff/project2-userlib"
	"github.com/google/uuid"

	// hex.EncodeToString(...) is useful for converting []byte to string

	// Useful for string manipulation
	//"strings"

	// Useful for formatting strings (e.g. `fmt.Sprintf`).
	"fmt"

	// Useful for creating new error messages to return using errors.New("...")
	"errors"

	// Optional.
	_ "strconv"
)

// This serves two purposes: it shows you a few useful primitives,
// and suppresses warnings for imports not being used. It can be
// safely deleted!
func someUsefulThings() {

	// Creates a random UUID.
	randomUUID := uuid.New()

	// Prints the UUID as a string. %v prints the value in a default format.
	// See https://pkg.go.dev/fmt#hdr-Printing for all Golang format string flags.
	userlib.DebugMsg("Random UUID: %v", randomUUID.String())

	// Creates a UUID deterministically, from a sequence of bytes.
	hash := userlib.Hash([]byte("user-structs/alice"))
	deterministicUUID, err := uuid.FromBytes(hash[:16])
	if err != nil {
		// Normally, we would `return err` here. But, since this function doesn't return anything,
		// we can just panic to terminate execution. ALWAYS, ALWAYS, ALWAYS check for errors! Your
		// code should have hundreds of "if err != nil { return err }" statements by the end of this
		// project. You probably want to avoid using panic statements in your own code.
		panic(errors.New("An error occurred while generating a UUID: " + err.Error()))
	}
	userlib.DebugMsg("Deterministic UUID: %v", deterministicUUID.String())

	// Declares a Course struct type, creates an instance of it, and marshals it into JSON.
	type Course struct {
		name      string
		professor []byte
	}

	course := Course{"CS 161", []byte("Nicholas Weaver")}
	courseBytes, err := json.Marshal(course)
	if err != nil {
		panic(err)
	}

	userlib.DebugMsg("Struct: %v", course)
	userlib.DebugMsg("JSON Data: %v", courseBytes)

	// Generate a random private/public keypair.
	// The "_" indicates that we don't check for the error case here.
	var pk userlib.PKEEncKey
	var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("PKE Key Pair: (%v, %v)", pk, sk)

	// Here's an example of how to use HBKDF to generate a new key from an input key.
	// Tip: generate a new key everywhere you possibly can! It's easier to generate new keys on the fly
	// instead of trying to think about all of the ways a key reuse attack could be performed. It's also easier to
	// store one key and derive multiple keys from that one key, rather than
	originalKey := userlib.RandomBytes(16)
	derivedKey, err := userlib.HashKDF(originalKey, []byte("mac-key"))
	if err != nil {
		panic(err)
	}
	userlib.DebugMsg("Original Key: %v", originalKey)
	userlib.DebugMsg("Derived Key: %v", derivedKey)

	// A couple of tips on converting between string and []byte:
	// To convert from string to []byte, use []byte("some-string-here")
	// To convert from []byte to string for debugging, use fmt.Sprintf("hello world: %s", some_byte_arr).
	// To convert from []byte to string for use in a hashmap, use hex.EncodeToString(some_byte_arr).
	// When frequently converting between []byte and string, just marshal and unmarshal the data.
	//
	// Read more: https://go.dev/blog/strings

	// Here's an example of string interpolation!
	_ = fmt.Sprintf("%s_%d", "file", 1)
}

// This is the type definition for the User struct.
// A Go struct is like a Python or Java class - it can have a			ttributes
// (e.g. like the Username attribute) and methods (e.g. like the StoreFile method below).
type User struct {
	Username string

	UserKey []byte //used later as a kay to encrypt files

	PrivateSigningKey userlib.DSSignKey

	PrivateKey userlib.PKEDecKey

	// You can add other attributes here if you want! But note that in order for attributes to
	// be included when this struct is serialized to/from JSON, they must be capitalized.
	// On the flipside, if you have an attribute that you want to be able to access from
	// this struct's methods, but you DON'T want that value to be included in the serialized value
	// of this struct that's stored in datastore, then you can use a "private" variable (e.g. one that
	// begins with a lowercase letter).
}

type Ciphertext struct {
	EncryptedMessage   []byte
	HMac               []byte
	HybridSymmetricKey []byte
}

type File struct {
	AccessControlUUID uuid.UUID

	AccessControlSpecificKey []byte

	IsOwner bool

	filename string
}

type OwnerAccessControl struct {
	FacUUID uuid.UUID 

	FacKey []byte

	OACcopies []SharedOAC
}

type FileAccessControl struct {
	StartChunkUUID uuid.UUID

	FileChunkSpecificKey []byte

	EndChunkUUID uuid.UUID
}

type SharedOAC struct {
	OACuuid uuid.UUID

	InviteUUID uuid.UUID

	OACKey []byte
}

type FileChunk struct {
	Content []byte

	NextChunkUUID uuid.UUID
}

type Invitation struct {
	SenderAccessControlUUID uuid.UUID
	SenderAccessControlKey  []byte
	//RecipientAccessControlUUID uuid.UUID
	SenderName        string
	RecipientUsername string
	FileName          string
}

// NOTE: The following methods have toy (insecure!) implementations.

////////////////////
//HELPER FUNCTIONS//
////////////////////

func setToDatastore(arg1 []byte, arg2 []byte, content []byte, id uuid.UUID, datatype string) (err error) {
	//we don't care about k1
	var uuidVar, _, k2, k3 []byte = getUuid(arg1, arg2)
	var uuidCast uuid.UUID
	if id != uuid.Nil {
		uuidCast = id
	} else {
		uuidCast, err = uuid.FromBytes(uuidVar[:16])
		if err != nil {
			return err
		}
	}

	encryptedMessage := userlib.SymEnc(k2[:16], userlib.RandomBytes(16), content)
	hMac, err3 := userlib.HMACEval(k3[:16], encryptedMessage)
	if err3 != nil {
		return err3
	}
	var ciphertext Ciphertext
	ciphertext.EncryptedMessage = encryptedMessage
	ciphertext.HMac = hMac
	ciphertextBytes, _ := json.Marshal(ciphertext)
	//prepare to put cipher text in data store
	userlib.DatastoreSet(uuidCast, ciphertextBytes)
	//userlib.DebugMsg("Updated UUID: %s (%s)\n", uuidCast, datatype)
	//just do panic
	return nil

}

func getFromDatastore(arg1 []byte, arg2 []byte, entryUUID uuid.UUID) (uuidCast uuid.UUID, content []byte, err error) {
	// K2, K3 = Get_UUID (arg1, arg2)
	// UUID = Hash(arg2)
	// If UUID is not specified, UUID = K2[0:15]
	// (EncryptedContent, HMAC) = Datastore.get(UUID)
	// If HMAC != Calculate HMAC (EncryptedContent)  => DATA INTEGRITY FAILED
	// Content = Decrypt(K2, EncryptedContent)   SymDec
	// Return Content

	var uuidVar, _, k2, k3 []byte = getUuid(arg1, arg2)

	uuidCast, _ = uuid.FromBytes(uuidVar[:16])
	if entryUUID != uuid.Nil {
		uuidCast = entryUUID
	}
	var ciphertext Ciphertext
	cipherTextBytes, ok := userlib.DatastoreGet(uuidCast)
	if !ok {
		return uuid.Nil, nil, errors.New("UUID not found")
	}
	err = json.Unmarshal(cipherTextBytes, &ciphertext)
	if err != nil {
		panic(errors.New("An error occurred while generating a UUID: " + err.Error()))

	}

	hMac, _ := userlib.HMACEval(k3[:16], ciphertext.EncryptedMessage)
	//check mac
	if !(userlib.HMACEqual(hMac, ciphertext.HMac)) {
		return uuid.Nil, nil, errors.New("not equal HMAC")

	}

	content = userlib.SymDec(k2[:16], ciphertext.EncryptedMessage)

	return uuidCast, content, nil
}

func getUuid(arg1 []byte, arg2 []byte) ([]byte, []byte, []byte, []byte) {

	var k1 []byte = userlib.Argon2Key(arg1, arg2, 16)
	var purposeK2 string = "encryption"
	purposeK2Bytes := []byte(purposeK2)
	k2, _ := userlib.HashKDF(k1, purposeK2Bytes)
	var purposeK3 string = "mac"
	purposeK3Bytes := []byte(purposeK3)
	k3, _ := userlib.HashKDF(k1, purposeK3Bytes)
	//hash based on userName
	var uuidBytes []byte = userlib.Hash(arg2)
	return uuidBytes, k1, k2, k3
}

func createFileStructure(fileName string, userKey []byte, accesControlKey []byte, accesControlUUID uuid.UUID, username string, isOwner bool) (fs File, err error) {
	//for storeFile, accessControlUUID is gonna be empty
	var file File
	file.AccessControlSpecificKey = accesControlKey
	file.AccessControlUUID = accesControlUUID
	file.IsOwner = isOwner
	file.filename = fileName

	// we don't need to do create an OAC if user isn't owner, copyOAC will do this and gives us the Access Control UUID
	if accesControlUUID == uuid.Nil {
		var uuidVar, _, _, _ []byte = getUuid(accesControlKey, accesControlKey)
		file.AccessControlUUID, _ = uuid.FromBytes(uuidVar[:16])
		createOwnerAccessControlStructure(username, file.AccessControlSpecificKey)
	}

	fileBytes, err := json.Marshal(file)
	if err != nil {
		return file, err
	}

	err = setToDatastore([]byte(fileName), []byte(string(userKey)+fileName), fileBytes, uuid.Nil, "file structure")
	if err != nil {
		return file, err
	}

	return file, nil

}

func copyOac(fileOacKey []byte, parentAccessUUID uuid.UUID, parentKey []byte) (err error) {
	var oac OwnerAccessControl
	oac.FacUUID = parentAccessUUID
	oac.FacKey= parentKey
	oacBytes, err := json.Marshal(oac)
	if err != nil {
		return err
	}
	err = setToDatastore(fileOacKey, fileOacKey, oacBytes, uuid.Nil, "copy OAC")
	if err != nil {
		return err
	}
	return nil

}

func createFileAccessControlStructure(username string, facKey []byte) (fac FileAccessControl, err error) {
	fac.StartChunkUUID = uuid.New()
	fac.FileChunkSpecificKey = userlib.RandomBytes(16)
	fac.EndChunkUUID = fac.StartChunkUUID //intially they point to the same uuid, when store/append is called end chunk uuid will move

	facBytes, err := json.Marshal(fac)
	if err != nil {
		return fac, err
	}

	err = setToDatastore(facKey, facKey, facBytes, uuid.Nil, "FAC")
	if err != nil {
		return fac, err
	}

	return fac, nil

}

func createOwnerAccessControlStructure(username string, fileOacKey []byte) (err error) {
	var oac OwnerAccessControl
	var facKey = userlib.RandomBytes(16)

	var uuidVar = userlib.Hash(facKey)
	var uuidCast uuid.UUID
	uuidCast, err = uuid.FromBytes(uuidVar[:16])
	if err != nil {
		return err
	}
	oac.FacUUID = uuidCast
	oac.FacKey = facKey
	_, err = createFileAccessControlStructure(username, facKey)
	if err != nil {
		return err
	}
	oacBytes, err := json.Marshal(oac)
	if err != nil {
		return err
	}

	err = setToDatastore(fileOacKey, fileOacKey, oacBytes, uuid.Nil, "OAC")
	if err != nil {
		return err
	}

	return nil

}

func getFAC(copyOAC OwnerAccessControl) (fac FileAccessControl, err error) { //getfilechunk
	var content []byte
	//userlib.DebugMsg("get fac parent uuid: %s, facKey:", copyOAC.FacUUID, copyOAC.FacKey)
	_, content, err = getFromDatastore(copyOAC.FacKey, copyOAC.FacKey, copyOAC.FacUUID)
	if err != nil {
		return fac, err
	}
	err = json.Unmarshal(content, &fac)
	if err != nil {
		return fac, err
	}
	return fac, nil
	

}

func getOAC(file File) (oac OwnerAccessControl, err error) {
	var content []byte
	_, content, err = getFromDatastore(file.AccessControlSpecificKey, file.AccessControlSpecificKey, uuid.Nil)
	if err != nil {
		return oac, err
	}
	err = json.Unmarshal(content, &oac)
	if err != nil {
		return oac, err
	}
	return oac, nil

}

func storeFileChunk(chunkUUID uuid.UUID, encKey []byte, content []byte) (endUUID uuid.UUID, err error) {
	//UUID, FILE_CHUNK_STRUCT = GET_FROM_DATASTORE(UUID, EncryptionKey, UUID)
	// If it already exists
	// 		Traverse the next UUID as list and delete all the UUIDs from datastore
	// Update the content in the File Chunk Struct to Content
	// Generate new Random UUID as the next pointer.
	// 		There is NO entry in the Data store with this UUID and so DataStore.get with this UUID will return ok = False indicating end of the list
	// SET_TO_DATASTORE(UUID, EncryptionKey, File Chunk Struct, UUID)

	var fcContent []byte
	var fc FileChunk
	_, fcContent, err = getFromDatastore(encKey, encKey, uuid.Nil)

	if err == nil {
		err = json.Unmarshal(fcContent, &fc)
		if err != nil {
			return uuid.Nil, err
		}

		// fc struct exists already in the datastore and we need to clean up dangling UUID (ones we no longer need) in the list
		//clean up here
	}

	fc.Content = content
	fc.NextChunkUUID = uuid.New()

	fcBytes, err := json.Marshal(fc)
	if err != nil {
		return uuid.Nil, err
	}

	err = setToDatastore(encKey, encKey, fcBytes, chunkUUID, "File Chunk")
	if err != nil {
		return uuid.Nil, err
	}
	//userlib.DebugMsg("chunk:%s, next chunk:%s, content:%s\n", chunkUUID, fc.NextChunkUUID, fc.Content)
	return fc.NextChunkUUID, nil
}

func deleteOACcopies(oac OwnerAccessControl, uuidCast uuid.UUID, oacKey []byte) (OwnerAccessControl, error) {
	//userlib.DebugMsg("OAC key:%s, uuidcast:%s, length of oacCopies:%d", oacKey, uuidCast, len(oac.OACcopies))
	var i int
	var childOAC OwnerAccessControl
	for i = 0; i < len(oac.OACcopies); i++ {
		if oac.OACcopies[i].InviteUUID == uuidCast || uuidCast == uuid.Nil { // either a specific uuid needs to be deleted or we are deleting everything
			_, childOACcontent, err := getFromDatastore(oac.OACcopies[i].OACKey, oac.OACcopies[i].OACKey, oac.OACcopies[i].OACuuid)
			if err != nil {
				return oac, err
			}
			err = json.Unmarshal(childOACcontent, &childOAC)
			if err != nil {
				return oac, err
			}
			deleteOACcopies(childOAC, uuid.Nil, oac.OACcopies[i].OACKey)
			//deleting the entries
			userlib.DatastoreDelete(oac.OACcopies[i].InviteUUID)
			userlib.DatastoreDelete(oac.OACcopies[i].OACuuid)
		}
		if oac.OACcopies[i].InviteUUID == uuidCast {
			break
		}
	}
	// cleaning out everything, not just updating one OAC
	if uuidCast != uuid.Nil {
		if i == len(oac.OACcopies) {
			return oac, errors.New("no invitation")
		}

		oac.OACcopies = append(oac.OACcopies[:i], oac.OACcopies[i+1:]...)
		oacBytes, err := json.Marshal(oac)
		if err != nil {
			return oac, err
		}
		err = setToDatastore(oacKey, oacKey, oacBytes, uuid.Nil, "OAC")
		if err != nil {
			return oac, err
		}
	}

	return oac, nil
}

func updateOACcopies(oac OwnerAccessControl) error {
	var i int
	for i = 0; i < len(oac.OACcopies); i++ {
		//userlib.DebugMsg("parent uuid: %s, updating oac copy: %s", oac.FacUUID, oac.OACcopies[i].OACuuid)
		err := copyOac(oac.OACcopies[i].OACKey, oac.FacUUID, oac.FacKey)
		if err != nil {
			return err
		}
		var content []byte
		_, content, err = getFromDatastore(oac.OACcopies[i].OACKey, oac.OACcopies[i].OACKey, uuid.Nil)
		if err != nil {
			return err
		}
		var childOAC OwnerAccessControl
		err = json.Unmarshal(content, &childOAC)
		if err != nil {
			return err
		}
		updateOACcopies(childOAC)
	}

	return nil
	
}

////////////////////
// Main Functions //
////////////////////

func InitUser(username string, password string) (userdataptr *User, err error) {
	if username == "" {
		return nil, errors.New("username is empty")
	}
	var uuidVar, _, _, _ []byte = getUuid([]byte(password), []byte(username))
	var uuidCast uuid.UUID
	uuidCast, err = uuid.FromBytes(uuidVar[:16])
	if err != nil {
		return nil, err
	}

	copy, _ := userlib.DatastoreGet(uuidCast)
	if copy != nil {
		return nil, errors.New("existing user")
	}
	var userdata User
	userdata.Username = username
	//userdata.UserKey = hex.EncodeToString(userlib.RandomBytes(16))
	//generate a random 16 bit key, which will be used to encrypt file structs. We convert it to a String for now

	//put username to datastore

	signKey, verifyKey, err := userlib.DSKeyGen()
	if err != nil {
		return nil, err
	}
	userdata.PrivateSigningKey = signKey
	userlib.KeystoreSet(userdata.Username+"sign", verifyKey)

	publicPKE, privatePKE, err := userlib.PKEKeyGen()
	if err != nil {
		return nil, err
	}
	userdata.PrivateKey = privatePKE
	userlib.KeystoreSet(userdata.Username+"PKE", publicPKE)

	userBytes, err := json.Marshal(userdata)
	if err != nil {
		return nil, err
	}

	err = setToDatastore([]byte(password), []byte(username), userBytes, uuid.Nil, "user")
	if err != nil {
		return nil, err
	}

	return &userdata, nil
}

func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata
	var content []byte
	_, content, err = getFromDatastore([]byte(password), []byte(username), uuid.Nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(content, userdataptr)
	if err != nil {
		return nil, err
	}
	return userdataptr, nil
}

// Check if the specified file is already present
// 		UUID, FS = GET_FROM_DATASTORE(filename, USER.username+USER.key)
// If file is not present, then create new File Structure
// 		FS = CreateFileStructure(filename, USER.username + USER.key, USER.username + USER.key)  This user will be the owner for the file.
// From the file structure FS,
// 		Get the StartingChunkUUID of the File (GetFileStartingChunk)
// 		StoreFileChunk(StartingChunkUUID, FS->AccessControl->encryptionKey, Content)

func (userdata *User) StoreFile(filename string, content []byte) (err error) {
	var file File
	var fsContent []byte
	repeat := false
	//check if that file already exists
	_, fsContent, err = getFromDatastore([]byte(filename), []byte(userdata.Username+string(userdata.UserKey)+filename), uuid.Nil)
	if err != nil {
		//File is not present, we need to create File Struct
		file, err = createFileStructure(filename, []byte(userdata.Username+string(userdata.UserKey)), userlib.RandomBytes(16), uuid.Nil, userdata.Username, true)
	} else {
		//update the content
		err = json.Unmarshal(fsContent, &file)
		repeat = true
	}
	if err != nil {
		return err
	}
	oac, err := getOAC(file)
	if err != nil {
		return err
	}
	fac, err := getFAC(oac)
	if err != nil {
		return err
	}
	if repeat {
		fac.EndChunkUUID = fac.StartChunkUUID
	}

	//userlib.DebugMsg("Store file before end uuid is changed : file start uuid: %s, file end uuid: %s, oac uuid: %s\n", fac.StartChunkUUID, fac.EndChunkUUID, file.AccessControlUUID)
	fac.EndChunkUUID, err = storeFileChunk(fac.EndChunkUUID, fac.FileChunkSpecificKey, content)
	if err != nil {
		return err
	}
	//userlib.DebugMsg("Store file, end uuid updated: file start uuid: %s, file end uuid: %s, oac uuid: %s\n", fac.StartChunkUUID, fac.EndChunkUUID, file.AccessControlUUID)
	facBytes, err := json.Marshal(fac)
	if err != nil {
		return err
	}
	
	err = setToDatastore(oac.FacKey, oac.FacKey, facBytes, uuid.Nil, "FAC")
	if err != nil {
		return err
	}

	return nil
}

func (userdata *User) AppendToFile(filename string, content []byte) error {
	var file File
	var fsContent []byte
	_, fsContent, err := getFromDatastore([]byte(filename), []byte(userdata.Username+string(userdata.UserKey)+filename), uuid.Nil)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fsContent, &file)
	if err != nil {
		return err
	}
	oac, err := getOAC(file)
	if err != nil {
		return err
	}
	fac, err := getFAC(oac)
	if err != nil {
		return err
	}
	//userlib.DebugMsg("Append file before end uuid is changed : file start uuid: %s, file end uuid: %s, oac uuid: %s\n", fac.StartChunkUUID, fac.EndChunkUUID, file.AccessControlUUID)
	fac.EndChunkUUID, err = storeFileChunk(fac.EndChunkUUID, fac.FileChunkSpecificKey, content)
	if err != nil {
		return err
	}
	//userlib.DebugMsg("Append file, end uuid updated: file start uuid: %s, file end uuid: %s, oac uuid: %s\n", fac.StartChunkUUID, fac.EndChunkUUID, file.AccessControlUUID)
	facBytes, err := json.Marshal(fac)
	if err != nil {
		return err
	}
	err = setToDatastore(oac.FacKey, oac.FacKey, facBytes, uuid.Nil, "FAC")
	if err != nil {
		return err
	}
	_, content, err = getFromDatastore(oac.FacKey, oac.FacKey, uuid.Nil)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &fac)
	if err != nil {
		return err
	}
	//userlib.DebugMsg("Append file, checking, end uuid updated: file start uuid: %s, file end uuid: %s, oac uuid: %s\n", fac.StartChunkUUID, fac.EndChunkUUID, file.AccessControlUUID)
	return nil
}

func (userdata *User) LoadFile(filename string) (content []byte, err error) {
	var file File
	var fsContent []byte
	_, fsContent, err = getFromDatastore([]byte(filename), []byte(userdata.Username+string(userdata.UserKey)+filename), uuid.Nil)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fsContent, &file)
	if err != nil {
		return nil, err
	}
	oac, err := getOAC(file)
	if err != nil {
		return nil, err
	}
	fac, err := getFAC(oac)
	if err != nil {
		return nil, err
	}
	//userlib.DebugMsg("Load file: file start uuid: %s, file end uuid: %s, oac uuid: %s\n", fac.StartChunkUUID, fac.EndChunkUUID, file.AccessControlUUID)

	chunkID := fac.StartChunkUUID
	for chunkID != fac.EndChunkUUID {
		var chunkContent FileChunk
		_, fcContent, err := getFromDatastore(fac.FileChunkSpecificKey, fac.FileChunkSpecificKey, chunkID)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(fcContent, &chunkContent)
		if err != nil {
			return nil, err
		}
		content = append(content, chunkContent.Content...)
		chunkID = chunkContent.NextChunkUUID
	}
	return content, nil
}

func (userdata *User) CreateInvitation(filename string, recipientUsername string) (invitationPtr uuid.UUID, err error) {
	var file File
	_, fs, err := getFromDatastore([]byte(filename), []byte(userdata.Username+string(userdata.UserKey)+filename), uuid.Nil)
	if err != nil {
		return uuid.Nil, err
	}
	err = json.Unmarshal(fs, &file)
	if err != nil {

		return uuid.Nil, err
	}

	newOACKey := userlib.RandomBytes(16)
	//get the owner OAC of the file
	oac, err := getOAC(file)
	if err != nil {
		return uuid.Nil, err
	}
	//create a copy OAC regardless
	copyOac(newOACKey, oac.FacUUID, oac.FacKey)

	var invite Invitation
	invite.SenderAccessControlKey = newOACKey
	invite.RecipientUsername = recipientUsername
	invite.SenderName = userdata.Username
	invite.FileName = filename
	var uuidVar = userlib.Hash(newOACKey)
	var uuidCast uuid.UUID
	uuidCast, err = uuid.FromBytes(uuidVar[:16])
	if err != nil {
		return uuid.Nil, err
	}
	invite.SenderAccessControlUUID = uuidCast
	//invite.RecipientAccessControlUUID = uuid.New()
	recipPublicKey, obtained := userlib.KeystoreGet(recipientUsername + "PKE")
	symKey := userlib.RandomBytes(16)
	encryptedSymKey, err := userlib.PKEEnc(recipPublicKey, symKey)
	if err != nil {
		return uuid.Nil, err
	}
	if !obtained {
		return uuid.Nil, errors.New("KeyStore")
	}
	inviteBytes, err := json.Marshal(invite)
	if err != nil {
		return uuid.Nil, err
	}
	encryptedMessage := userlib.SymEnc(symKey, userlib.RandomBytes(16), inviteBytes)
	if err != nil {
		return uuid.Nil, err
	}
	signed, err := userlib.DSSign(userdata.PrivateSigningKey, encryptedMessage)
	if err != nil {
		return uuid.Nil, err
	}
	inviteUUID := userlib.Hash([]byte(userdata.Username + recipientUsername + filename))
	uuidCast, err = uuid.FromBytes(inviteUUID[:16])
	if err != nil {
		return uuid.Nil, err
	}
	var ciphertext Ciphertext
	ciphertext.EncryptedMessage = encryptedMessage
	ciphertext.HMac = signed
	ciphertext.HybridSymmetricKey = encryptedSymKey
	ciphertextBytes, err := json.Marshal(ciphertext)
	if err != nil {
		return uuid.Nil, err
	}
	userlib.DatastoreSet(uuidCast, ciphertextBytes)

	var sharedoac SharedOAC
	sharedoac.InviteUUID = uuidCast
	sharedoac.OACuuid = invite.SenderAccessControlUUID
	sharedoac.OACKey = newOACKey

	oac, err = getOAC(file)
	if err != nil {
		return uuid.Nil, err
	}
	oac.OACcopies = append(oac.OACcopies, sharedoac)
	oacBytes, err := json.Marshal(oac)
	if err != nil {
		return uuid.Nil, err
	}
	err = setToDatastore(file.AccessControlSpecificKey, file.AccessControlSpecificKey, oacBytes, uuid.Nil, "OAC")
	if err != nil {
		return uuid.Nil, err
	}
	return uuidCast, nil
}

func (userdata *User) AcceptInvitation(senderUsername string, invitationPtr uuid.UUID, filename string) error {

	//recreate the invitation UUID
	//inviteUUID := userlib.Hash([]byte(senderUsername + userdata.Username + filename))
	inviteUUID := invitationPtr
	//uuidCast, err := uuid.FromBytes(inviteUUID[:16])

	//Get the cipher text that contains the invitation UUID
	cipherInviteBytes, ok := userlib.DatastoreGet(inviteUUID)
	if !ok {
		return errors.New("DataStoreGet didn't return ok")
	}
	var ciphertext Ciphertext
	err := json.Unmarshal(cipherInviteBytes, &ciphertext)
	if err != nil {
		return err
	}
	senderVerifyKey, ok := userlib.KeystoreGet(senderUsername + "sign")
	if !ok {
		return errors.New("keystore get not ok")
	}
	signedBytes := ciphertext.HMac
	err = userlib.DSVerify(senderVerifyKey, ciphertext.EncryptedMessage, signedBytes)
	if err != nil {
		return err
	}
	//signature is verified, so lets decrypt our message
	//the symmetric key was encrypted with the recippublic key, so lets first decrypt with the recipprivate key
	symmKey, err := userlib.PKEDec(userdata.PrivateKey, ciphertext.HybridSymmetricKey)
	if err != nil {
		return err
	}
	inviteBytes := userlib.SymDec(symmKey, ciphertext.EncryptedMessage)
	var invite Invitation
	err = json.Unmarshal(inviteBytes, &invite)
	if err != nil {
		return err
	}
	createFileStructure(filename, []byte(userdata.Username+string(userdata.UserKey)), invite.SenderAccessControlKey, invite.SenderAccessControlUUID, userdata.Username, false)

	return nil
}

func (userdata *User) RevokeAccess(filename string, recipientUsername string) error {
	// Get the specified file
	// 	UUID, FS = GET_FROM_DATASTORE(filename, USER.username + USER.key)
	// Invitation_UUID, INVITATION_STRUCT = GET_FROM_DATASTORE(USER.username + USER.key, recipient)
	// Delete the recipient access control UUID in the invitation structure in DataStore
	// Delete Invitation_UUID in Datastore
	// Change the file specific key for the actual chunk and encrypt the file chunk again with a new key.
	// Also change the UUID for OWNER_ACCESS_CONTROL_STRUCT. This will prevent someone with the old information saved from being able to get the file.

	var file File
	var fsContent []byte
	_, fsContent, err := getFromDatastore([]byte(filename), []byte(userdata.Username+string(userdata.UserKey)+filename), uuid.Nil)
	if err != nil {
		return err
	}
	err = json.Unmarshal(fsContent, &file)
	if err != nil {
		return err
	}

	inviteUUID := userlib.Hash([]byte(userdata.Username + recipientUsername + filename))
	uuidCast, err := uuid.FromBytes(inviteUUID[:16])
	if err != nil {
		return err
	}

	oac, err := getOAC(file)
	if err != nil {
		return err
	}
	fac, err := getFAC(oac)
	

	if err != nil {
		return err
	}

	//update FAC
	
	//generate new FAC Key
	newFACKey := userlib.RandomBytes(16)
	newFAC, err := createFileAccessControlStructure(userdata.Username, newFACKey)
	if err != nil {
		return err
	}
	var content []byte
	chunkID := fac.StartChunkUUID
	for chunkID != fac.EndChunkUUID {
		var chunkContent FileChunk
		_, fcContent, err := getFromDatastore(fac.FileChunkSpecificKey, fac.FileChunkSpecificKey, chunkID)
		if err != nil {
			return err
		}
		err = json.Unmarshal(fcContent, &chunkContent)
		if err != nil {
			return err
		}
		content = append(content, chunkContent.Content...)
		chunkID = chunkContent.NextChunkUUID
	}
	newFAC.EndChunkUUID, err = storeFileChunk(newFAC.EndChunkUUID, newFAC.FileChunkSpecificKey, content)
	if err != nil {
		return err
	}
	//userlib.DebugMsg("Storing new copy of file after revoke is called: file start uuid: %s, file end uuid: %s, oac uuid: %s\n", newFAC.StartChunkUUID, newFAC.EndChunkUUID, file.AccessControlUUID)
	facBytes, err := json.Marshal(newFAC)
	if err != nil {
		return err
	}
	
	err = setToDatastore(newFACKey, newFACKey, facBytes, uuid.Nil, "FAC")
	if err != nil {
		return err
	}

	//update OAC to delete shared
	oac, err = deleteOACcopies(oac, uuidCast, file.AccessControlSpecificKey)
	if err != nil {
		return err
	}


	//update OAC to update the shared to point to the new FAC
	oac.FacKey = newFACKey
	var uuidVar, _, _, _ []byte = getUuid(newFACKey, newFACKey)
	uuidCast, err = uuid.FromBytes(uuidVar[:16])
	if err != nil {
		return err
	}
	oac.FacUUID = uuidCast
	//userlib.DebugMsg("revoke is called here, updating oacs, new face uuid:%s", oac.FacUUID)
	updateOACcopies(oac)
	oacBytes, err := json.Marshal(oac)

	if err != nil {
		return err
	}

	err = setToDatastore(file.AccessControlSpecificKey, file.AccessControlSpecificKey, oacBytes, uuid.Nil, "OAC")
	if err != nil {
		return err
	}

	return nil
}
