# ItemObject

ActorObjects are the basic object that represents Users and NPCs

- [ItemObject](#itemobject)
  - [CreateItem(itemId int) ItemObject ](#createitemitemid-int-itemobject-)
  - [ItemObject.ItemId() int](#itemobjectitemid-int)
  - [ItemObject.GetUsesLeft() int](#itemobjectgetusesleft-int)
  - [ItemObject.SetUsesLeft(amount int) int](#itemobjectsetusesleftamount-int-int)
  - [ItemObject.AddUsesLeft(amount int) int](#itemobjectaddusesleftamount-int-int)
  - [ItemObject.GetLastUsedRound() uint64](#itemobjectgetlastusedround-uint64)
  - [ItemObject.MarkLastUsed(clear bool) uint64](#itemobjectmarklastusedclear-bool-uint64)
  - [ItemObject.DisplayName( \[plainFormat bool\] ) string](#itemobjectdisplayname-plainformat-bool--string)
  - [ItemObject.NameSimple() string](#itemobjectnamesimple-string)
  - [ItemObject.NameComplex() string](#itemobjectnamecomplex-string)
  - [ItemObject.SetTempData(key string, value any)](#itemobjectsettempdatakey-string-value-any)
  - [ItemObject.GetTempData(key string) any](#itemobjectgettempdatakey-string-any)
  - [ItemObject.Rename(newName string \[, displayNameOrStyle string\])](#itemobjectrenamenewname-string--displaynameorstyle-string)
  - [ItemObject.Redescribe(newDescription string)](#itemobjectredescribenewdescription-string)

## [CreateItem(itemId int) ItemObject ](/scripting/item_func.go)
Creates a new instance of an item and returns it.

|  Argument | Explanation |
| --- | --- |
| itemId | The item id to create an instance of. |

## [ItemObject.ItemId() int](/scripting/item_func.go)
Returns the itemId of the ItemObject.

## [ItemObject.GetUsesLeft() int](/scripting/item_func.go)
Returns the number of uses remaining on the item (if any).

## [ItemObject.SetUsesLeft(amount int) int](/scripting/item_func.go)
Sets the remaining uses for the item to a specific number.

|  Argument | Explanation |
| --- | --- |
| amount | The number of uses to set the item to. |

## [ItemObject.AddUsesLeft(amount int) int](/scripting/item_func.go)
Adds a positive or negative quantity of uses to the item.

|  Argument | Explanation |
| --- | --- |
| amount | Positive of Negative number to add. |

## [ItemObject.GetLastUsedRound() uint64](/scripting/item_func.go)
Gets the last round number the item was used.

## [ItemObject.MarkLastUsed(clear bool) uint64](/scripting/item_func.go)
Set the last used round to the current round, or optionally clear it.

|  Argument | Explanation |
| --- | --- |
| clear (optional) | If true, will clear the last used back to zero |

## [ItemObject.DisplayName( [plainFormat bool] ) string](/scripting/item_func.go)
Returns the name of the object, such as "Glowing Battleaxe"

|  Argument | Explanation |
| --- | --- |
| plainFormat (optional) | If true, will provide plain text name without special colors. |

## [ItemObject.NameSimple() string](/scripting/item_func.go)
Returns the simple name of the object. For example, a "Glowing Battleaxe" may just be "Axe"

## [ItemObject.NameComplex() string](/scripting/item_func.go)
Returns the complex name of the object, such as "Glowing Batteaxe +2 [c]"

## [ItemObject.SetTempData(key string, value any)](/scripting/item_func.go)
Sets temporary data of any sort on the item. This data is not saved/loaded when despawning.

|  Argument | Explanation |
| --- | --- |
| key | The name to store the data under. Also used to retrieve the data later. |
| vaue | The data to store. |

## [ItemObject.GetTempData(key string) any](/scripting/item_func.go)
Sets temporary data of any sort on the item. This data is not saved/loaded when despawning.

|  Argument | Explanation |
| --- | --- |
| key | The name to retrieve data under. |


## [ItemObject.Rename(newName string [, displayNameOrStyle string])](/scripting/item_func.go)
Renames the item, also optionally provide a fancy name or colorpattern

|  Argument | Explanation |
| --- | --- |
| newName | The plaintext name. |
| displayNameOrStyle | A fancy name in ansi tags, color short tags, or a pattern like :flame |

## [ItemObject.Redescribe(newDescription string)](/scripting/item_func.go)
Change the description for an item

|  Argument | Explanation |
| --- | --- |
| newDescription | The plaintext new description. |
