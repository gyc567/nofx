# Bug Fix: Web3Auth Signature Hex Encoding

## Issue
Signature recovery tests in `nofx/web3_auth/signatures_test.go` were failing with "hex string without 0x prefix" error.

## Root Cause Analysis

### Problem Statement
The `signMessage()` test helper returned signatures without the "0x" prefix:
```go
return hex.EncodeToString(signature)  // ❌ Returns "a1b2c3..."
```

But `RecoverAddressFromSignature()` expected signatures with "0x" prefix:
```go
sigBytes, err := hexutil.Decode(signature)  // Expects "0xa1b2c3..."
```

### Architecture Issue
Inconsistent format expectations between signing and recovery functions:
- **Signing**: Produced bare hex string
- **Recovery**: Required 0x-prefixed hex string

This is a common pattern mismatch in Ethereum libraries where:
- Internal functions often work with bare hex
- Public APIs expect 0x prefix (EIP-55 standard)

### Philosophy
Ethereum's de facto standard (EIP-55) requires "0x" prefix for all hex-encoded values. The recovery function was correct; the test helper was incomplete.

## Solution

### Fix Description
Updated `signMessage()` to prepend "0x" prefix before returning:

```go
// Before
return hex.EncodeToString(signature), nil

// After
return "0x" + hex.EncodeToString(signature), nil
```

### Test Coverage
Now properly validates the entire signature flow:
1. Generate key pair → Address
2. Sign message with EIP-191 standard
3. Recover address from signature
4. Verify recovered == original

### Code Changes
File: `nofx/web3_auth/signatures_test.go:41`

```go
func signMessage(privateKey *ecdsa.PrivateKey, message string) (string, error) {
    msgHash := generateMessageHash(message)
    signature, err := crypto.Sign(msgHash, privateKey)
    if err != nil {
        return "", err
    }
    return "0x" + hex.EncodeToString(signature), nil  // ✅ Added prefix
}
```

## Test Results
✅ All web3_auth tests now pass (8/8)
- TestRecoverAddressFromSignature_Valid ✅
- TestRecoverAddressFromSignature_InvalidSignature ✅
- TestRecoverAddressFromSignature_AddressMismatch ✅
- TestValidateAddress ✅
- TestValidateSignature ✅
- TestValidateNonce ✅
- TestGenerateSignatureMessage ✅
- TestHashSignature ✅
- TestSanitizeAddress ✅

## Files Modified
- `nofx/web3_auth/signatures_test.go:41`

## Impact
- ✅ **Production**: No changes to main implementation
- ✅ **Testing**: Proper end-to-end signature verification
- ✅ **Standard Compliance**: Aligns with EIP-55 hex encoding standard

## Category
Test Utility / Format Consistency / Ethereum Standard Compliance
