# Secure File Storage and Sharing

This repository implements a secure file storage and sharing system designed for CS161, focusing on cryptographic techniques and secure data handling. The package allows users to initialize accounts, store and retrieve files securely, share files with other users, and revoke access when necessary.

Implementation is in `client/client.go` and integration test cases are `client_test/client_test.go`. Other additional unit test cases provided in `client/client_unittest.go`. For comprehensive documentation, see the https://cs161.org/proj2/.

Features
 - User Initialization: Create a new user with secure credentials and encryption keys.
 - File Storage and Retrieval: Store files securely with encrypted content and integrity checks.
 - File Sharing: Share access to files using cryptographic invitations that allow secure, controlled file access.
 - Access Revocation: Revoke access to files from specific users, ensuring previous keys are no longer valid.

Project Structure
 - User struct: Manages user identity and encryption keys.
 - File struct: Handles file metadata and access control.
 - Helper functions and methods provide secure data storage, retrieval, and sharing capabilities, relying on cryptographic primitives such as UUIDs, encryption keys, and digital signatures.

Usage
This project was developed with limited dependencies, restricted to cryptographic libraries provided by the CS161 staff to ensure security and autograder compatibility.
