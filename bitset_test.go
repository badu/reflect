package reflector

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"
)

func TestString(t *testing.T) {
	role := &UserRole{
		Permisions: New(),
	}

	role.Permisions.SetAll(Login, ListSchedulesForAllUsers, ContactList, GreetingsList)
	role.Rights = base64.StdEncoding.EncodeToString(role.Permisions)

	t.Logf("Encoded right : %q", role.Rights)

	stringRights, err := base64.StdEncoding.DecodeString("AgAQAAABAAAAAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAA")
	if err != nil {
		t.Fatalf("Error decoding string : %v", err)
	}

	otherRole := &UserRole{}
	otherRole.Permisions.Load(stringRights)

	if ok, err := otherRole.Permisions.Test(Login); err != nil || !ok {
		for i, aByte := range otherRole.Permisions {
			t.Logf("%d %08b\n", i, byte(aByte))
			for j := uint(0); j < uint(8); j++ {
				t.Logf("%d %d -> %t", i, j, (aByte&(1<<j) != 0))
			}
		}
		t.Fatalf("Error on test all.")
	}
	if ok, err := otherRole.Permisions.Test(ListSchedulesForAllUsers); err != nil || !ok {
		for i, aByte := range otherRole.Permisions {
			t.Logf("%d %08b\n", i, byte(aByte))
			for j := uint(0); j < uint(8); j++ {
				t.Logf("%d %d -> %t", i, j, (aByte&(1<<j) != 0))
			}
		}
		t.Fatalf("Error on test all.")
	}
	if ok, err := otherRole.Permisions.Test(ContactList); err != nil || !ok {
		for i, aByte := range otherRole.Permisions {
			t.Logf("%d %08b\n", i, byte(aByte))
			for j := uint(0); j < uint(8); j++ {
				t.Logf("%d %d -> %t", i, j, (aByte&(1<<j) != 0))
			}
		}
		t.Fatalf("Error on test all.")
	}
	if ok, err := otherRole.Permisions.Test(GreetingsList); err != nil || !ok {
		for i, aByte := range otherRole.Permisions {
			t.Logf("%d %08b\n", i, byte(aByte))
			for j := uint(0); j < uint(8); j++ {
				t.Logf("%d %d -> %t", i, j, (aByte&(1<<j) != 0))
			}
		}
		t.Fatalf("Error on test all.")
	}
}

func randomValuesSet() ([]byte, []uint64) {
	var whichBits []uint64

	//searching the maximum number of bytes needed for this slice
	maxRightId := uint64(0)
	for _, right := range Rights() {
		if uint64(right) > maxRightId {
			maxRightId = uint64(right)
		}
	}
	size := (maxRightId >> 3) + 1
	fmt.Printf("Max right value : %d [%d] slice of bytes of len %d\n", maxRightId, maxRightId>>3, size)
	//creating a slice and pretending it came from database
	ret := make([]byte, size)
	whichBits = make([]uint64, 0)
	for i := uint64(0); i < size; i++ {
		for j := uint(0); j < uint(8); j++ {
			//having some random permissions
			if rand.Intn(2) == 1 {
				//fmt.Printf("      set %d %d %08b\n", i, j, ret[i])
				ret[i] = ret[i] | (1 << j)
				where := i*8 + uint64(j)
				whichBits = append(whichBits, where)
				//fmt.Printf("after set %d %d %08b\n", i, j, ret[i])
			}
		}
	}
	return ret, whichBits
}

func listRightBitSetContent(t *testing.T, rightBitSet *BitSet) {
	fmt.Printf("Listing %d bytes.", rightBitSet.Len())
	for i := uint64(0); i < uint64(rightBitSet.Len()*8); i++ {
		if ok, err := rightBitSet.Test(BitIndex(i)); err != nil || !ok {
			found := false
			byteNo := i >> 3
			bitNo := i - byteNo*8
			for _, right := range Rights() {
				if uint64(right) == i {
					//t.Logf("bit #%d is set (byte = %d bit = %d)\n", i, byteNo, bitNo)
					//t.Logf("RIGHT #%d : %s\n", right, right.Name())
					found = true
				}
			}
			if !found {
				t.Logf("Constant not declared for bit #%d is set (byte = %d bit = %d)\n", i, byteNo, bitNo)
			}
		}
	}
}

func TestEverything(t *testing.T) {
	t.Log("Randomizing permission flags...")
	testData, whichBits := randomValuesSet()

	for i, aByte := range testData {
		t.Logf("byte %d value %08b\n", i, byte(aByte))
	}
	t.Logf("Number of rights = %d => minimum slice size %d\n", len(Rights()), len(testData))

	rbs, err := MakeFromByteArray(testData)
	if err != nil {
		t.Fatalf("Error : %v", err)
	}

	for _, bit := range whichBits {
		t.Logf("Setting Bit %d\n", bit)
	}

	listRightBitSetContent(t, &rbs)

	if ok, err := rbs.Test(ContactDelete); err != nil || !ok {
		t.Logf("Allowing rightContactDelete\n")
		rbs.Set(ContactDelete)
	}

	if ok, err := rbs.Test(ContactDelete); err != nil || !ok {
		t.Fatalf("ERROR : Should have right to delete contact\n")
	}

	rbs.Unset(ContactDelete)

	if ok, err := rbs.Test(ContactDelete); err != nil || ok {
		t.Fatalf("ERROR : Should NOT have right to delete contact\n")
	}

	t.Logf("%s\n", rbs.String())
	t.Logf("Content of the %d byte slice (binary):\n", rbs.Len())
	for i, aByte := range rbs {
		t.Logf("%d %08b\n", i, byte(aByte))
		for j := uint(0); j < uint(8); j++ {
			t.Logf("%d %d -> %t", i, j, (aByte&(1<<j) != 0))
		}
	}
}

func TestFromNew(t *testing.T) {
	rbs := New()
	t.Logf("TEST SET\n")
	rbs.SetAll(ContactAdd, ContactEdit, ContactDelete, Login)
	listRightBitSetContent(t, &rbs)
	t.Logf("TEST UNSET")
	rbs.UnsetAll(ContactAdd, ContactEdit, ContactDelete)
	listRightBitSetContent(t, &rbs)
	/**
	for i, aByte := range rbs {
		t.Logf("%d %08b\n", i, byte(aByte))
		for j := uint(0); j < uint(8); j++ {
			t.Logf("%d %d -> %t", i, j, (aByte & (1 << j) != 0))
		}
	}
	**/
}
