# Warehouse Design Implementation Guide

This document explains the Go implementation of the warehouse-store distance model. The design uses explicit StoreID maps for type safety, clarity, and extensibility.

## 1. Data Models
**File:** `models/warehouse_structure.go`

This file defines the core data structures used to represent the physical network.

- **`StoreNode`**: Represents a brick-and-mortar store.
  - We use a dedicated struct (`StoreNode` vs `Store`) to decouple this logic from the rest of the application during the refactoring phase.
  - Contains `ID` (int) and `Name` (string).

- **`WarehouseNode`**: Represents a warehouse facility.
  - **`Distances`**: A map of strings to integers (`map[string]int`) where the key is the unique `StoreID`.
  - Example: If `Distances` is `{"store-a": 5, "store-b": 12, "store-c": 3}`, it means:
    - 5km to Store A
    - 12km to Store B
    - 3km to Store C
  - **Benefits**:
    - Eliminates "index out of bounds" errors
    - Doesn't rely on strict list ordering
    - Self-documenting (explicit store IDs)
    - Easy to add/remove stores without breaking existing data

**File:** `models/collectible.go`

- **`Warehouse`**: Represents a storage location for collectibles.
  - Uses `Distances map[string]int` for StoreID → distance mapping
  - All seed data uses explicit StoreID keys: `"store-a"`, `"store-b"`, `"store-c"`

## 2. Business Logic & Validation
**File:** `services/warehouse_manager.go`

This service manages the lifecycle and rules of the warehouse network.

### Validation

Two key validation functions are enforced:

1. **`ValidateStoreCount`**: Ensures there are always at least 3 stores active. This prevents edge cases in tri-location logic.

2. **`ValidateDistanceMap`**: Ensures all active stores have valid distance entries.
   - Checks that every store in the system has a distance entry in the warehouse's map
   - Validates that all distances are non-negative
   - Fails fast if a store ID is missing or has invalid data

### Construction

- **`ConvertDistances`**: A utility that transforms legacy distance tuples (`[]int`) into the `Distances` map using the master store definitions. This is maintained for backward compatibility with external data sources.

- **`BuildWarehouses`**: This function acts as a factory. It takes raw tuples, converts them via `ConvertDistances`, validates the resulting maps, and produces safe `WarehouseNode` objects.

### Allocation Strategy

- **`FindNearestWarehouse`**: This function determines which warehouse should fulfill an order for a specific store.
  - **Input**: The list of all warehouses and the target `storeID` (string).
  - **Logic**: It performs a direct map lookup (`w.Distances[storeID]`). If a warehouse serves that store, its distance is compared against the current minimum.
  - **Output**: Returns the `WarehouseNode` with the mathematically lowest distance to the requested store.
  - **Error Handling**: Returns an error if no warehouse has a distance entry for the requested store.

## 3. Allocation Manager
**File:** `services/allocation_manager.go`

The `AllocationManager` handles collectible allocation and ETA calculation using the map-based distance system.

### Key Methods

- **`Allocate(collectibleID, storeID)`**: Finds and allocates the nearest available unit.
  - Uses `warehouse.Distances[storeID]` for distance lookups
  - Filters available units by collectible ID
  - Selects the warehouse with minimum distance
  - Marks the unit as unavailable after allocation

- **`GetETA(collectibleID, storeID)`**: Calculates estimated delivery time.
  - Returns the minimum distance across all available units
  - Uses map lookups exclusively (no index-based logic)

## 4. Data Seeding
**File:** `data/seed.go`

All warehouse seed data uses explicit StoreID maps:

```go
repo.AddWarehouse("col-001", models.Warehouse{
    ID:            "wh-001-1",
    Name:          "Warehouse North - Batman",
    CollectibleID: "col-001",
    Available:     true,
    Distances: map[string]int{
        "store-a": 1,
        "store-b": 4,
        "store-c": 5,
    },
})
```

This format is:
- **Self-documenting**: Store IDs are explicit
- **Order-independent**: No reliance on array indices
- **Extensible**: New stores can be added without modifying existing data

## 5. Testing
**File:** `services/warehouse_manager_test.go`

The test suite verifies that the rules above are strictly enforced.

### Edge Cases Tested

- **Store Count Validation**: Providing fewer than 3 stores (should fail)
- **Distance Map Validation**:
  - Missing store distance entries (should fail)
  - Negative distance values (should fail)
- **Allocation Logic**:
  - Requesting allocation for unknown store IDs (should fail gracefully)
  - Verifying nearest warehouse selection with map-based lookups

### Happy Path

- Verifies that given valid inputs, the system correctly identifies the mathematically closest warehouse
- Confirms that distance lookups work correctly with explicit StoreID keys
- Validates that allocation behavior matches expected results

## 6. Migration Notes

### What Changed

- **Removed**: All tuple-based distance logic (`DistancesToStores []int`)
- **Added**: Map-based distances (`Distances map[string]int`)
- **Removed**: Legacy `AllocationService` that used index-based lookups
- **Enhanced**: Validation now checks for missing stores and negative distances

### Backward Compatibility

The `ConvertDistances` utility function remains available for importing legacy data from external sources that still use tuple format. However, all internal business logic uses map-based lookups exclusively.

### Benefits

1. **Type Safety**: StoreID-based maps eliminate "index out of bounds" errors
2. **Clarity**: Explicit store IDs make code self-documenting
3. **Extensibility**: Adding new stores doesn't require reordering or index management
4. **Maintainability**: No hidden dependencies on store ordering
5. **Validation**: Can validate that all required stores have distance entries
