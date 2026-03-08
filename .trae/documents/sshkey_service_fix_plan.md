# Git Manage Service - SSH Key Service Fix Plan

## \[x] Task 1: Fix sshkey\_service.go imports and response format

* **Priority**: P0

* **Depends On**: None

* **Description**:

  * Add necessary imports for db, response, and other required packages

  * Update response format to use the统一的 response package

* **Success Criteria**:

  * All necessary imports are added

  * Response format is consistent with other service files

* **Test Requirements**:

  * `programmatic` TR-1.1: No compilation errors

  * `human-judgement` TR-1.2: Imports are properly organized and response format is consistent

## \[x] Task 2: Implement ListDBSSHKeys function

* **Priority**: P0

* **Depends On**: Task 1

* **Description**:

  * Implement the ListDBSSHKeys function to fetch all SSH keys from the database

  * Convert database models to response format

* **Success Criteria**:

  * Function returns all SSH keys from the database

  * Response format is correct

* **Test Requirements**:

  * `programmatic` TR-2.1: Function returns 200 status with correct data

  * `programmatic` TR-2.2: Empty list is returned when no keys exist

## \[x] Task 3: Implement CreateDBSSHKey function

* **Priority**: P0

* **Depends On**: Task 1

* **Description**:

  * Implement the CreateDBSSHKey function to create a new SSH key in the database

  * Validate input data

  * Check for duplicate names

* **Success Criteria**:

  * New SSH key is created in the database

  * Response returns the created key with correct data

* **Test Requirements**:

  * `programmatic` TR-3.1: Function returns 200 status with created key data

  * `programmatic` TR-3.2: Duplicate name returns 400 status

  * `programmatic` TR-3.3: Invalid input returns 400 status

## \[x] Task 4: Implement GetDBSSHKey function

* **Priority**: P0

* **Depends On**: Task 1

* **Description**:

  * Implement the GetDBSSHKey function to fetch a single SSH key by ID

  * Handle not found case

* **Success Criteria**:

  * Function returns the SSH key with the specified ID

  * Not found case returns appropriate error

* **Test Requirements**:

  * `programmatic` TR-4.1: Function returns 200 status with correct key data

  * `programmatic` TR-4.2: Non-existent ID returns 404 status

## \[x] Task 5: Implement UpdateDBSSHKey function

* **Priority**: P0

* **Depends On**: Task 1

* **Description**:

  * Implement the UpdateDBSSHKey function to update an existing SSH key

  * Validate input data

  * Check for duplicate names (excluding current key)

* **Success Criteria**:

  * SSH key is updated in the database

  * Response returns the updated key with correct data

* **Test Requirements**:

  * `programmatic` TR-5.1: Function returns 200 status with updated key data

  * `programmatic` TR-5.2: Duplicate name returns 400 status

  * `programmatic` TR-5.3: Non-existent ID returns 404 status

## \[x] Task 6: Implement DeleteDBSSHKey function

* **Priority**: P0

* **Depends On**: Task 1

* **Description**:

  * Implement the DeleteDBSSHKey function to delete an SSH key by ID

  * Handle not found case

* **Success Criteria**:

  * SSH key is deleted from the database

  * Response returns success message

* **Test Requirements**:

  * `programmatic` TR-6.1: Function returns 200 status with success message

  * `programmatic` TR-6.2: Non-existent ID returns 404 status

## \[x] Task 7: Implement TestDBSSHKey function

* **Priority**: P0

* **Depends On**: Task 1

* **Description**:

  * Implement the TestDBSSHKey function to test SSH key connection

  * Use git service to test the connection

* **Success Criteria**:

  * Function tests the SSH key connection and returns the result

* **Test Requirements**:

  * `programmatic` TR-7.1: Function returns 200 status with test result

  * `programmatic` TR-7.2: Non-existent ID returns 404 status

## \[x] Task 8: Run tests and verify all functions work correctly

* **Priority**: P1

* **Depends On**: All previous tasks

* **Description**:

  * Run any available tests

  * Verify all functions work correctly

* **Success Criteria**:

  * All tests pass

  * All functions work as expected

* **Test Requirements**:

  * `programmatic` TR-8.1: No compilation errors

  * `programmatic` TR-8.2: All functions return correct status codes

