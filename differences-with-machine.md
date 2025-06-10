The implementation of the numscript interpreter behaves mostly the same with the machine implementation that is embedded in the [ledger](https://github.com/formancehq/ledger) repo, however there are new functionalities and some breaking changes that are important to be aware of.

# Breaking changes

### Zero postings are trimmed

The interpreter version never emits postings with zero amount anymore.
For example, the following numscript:

```numscript
send [USD/2 0] (
  source = @alice
  destination = @bob
)
```

resulted in the `[{source: "alice", destination: "bob", amount: 0, asset: "USD/2"}]` posting in the machine, while now it doesn't output postings anymore.
This is true for **every** kind of posting with a zero amount, which now aren't outputted anymore, with no exceptions.

<details>
<summary>Rationale</summary>
The previous behaviour was an edge case that wasn't explicitly documented and didn't have clear rules, which could make the users rely on undefined behaviours.

Sometimes this led to confusing scenarios:

```numscript
send [USD/2 0] (
  source = {
    1/2 from @s1
    1/2 from @s2
  }
  destination = @world
)
```

which would generate those postings:
`[{source: "s1", destination: "world", amount: 0, asset: "USD/2"}]` (no `@s2`)

Sometimes it led to noisy outputs:

```numscript
// if @alice had an empty balance, this was previously a zero posting
// now no postings are emitted
send [USD/2 100] (
  source = {
    @s1
    @s2
    @s3
    @s4
  }
  destination = @world
)
```

Possible postings:

```
[
  {source: "s1", destination: "world", amount: 0, asset: "USD/2"},
  {source: "s2", destination: "world", amount: 0, asset: "USD/2"},
  {source: "s3", destination: "world", amount: 0, asset: "USD/2"},
  {source: "s4", destination: "world", amount: 100, asset: "USD/2"},
]
```

</details>

# New functionalities

### Newlines are now optional

Missing newlines won't be a parsing error anymore. For example, the following is a valid numscript:

```numscript
send [USD/2 100] (
  source = { @a @b }
  destination = { remaining kept }
)
```

### Optional underscore in numbers literal

You can now write numeric literals with underscores. They only act as visual separators and have no semantic meaning:

```java
123_456 // equivalent to 123456

// underscores can be applied anywhere and group
// by any number of digits
1_0_20_50_0_2
```

You cannot use underscore in leading or trailing position, and there cannot be many underscores in a row. For example, all the following are forbidden:

```
_100
100_
1__00 // <- only one _ in a row allowed
```

### Parenthesis in expressions

You can now group expressions by parenthesis. For example, expressions like this are now possible:

```numscript
10 - ($n + $m)
```

### Proper infix `/` operator

The division operator is now a proper infix operator (not just syntax for the portion literal), of type `(number, number) -> portion`.
This means you can write more flexible expressions:

```numscript
$n/$m
1/($n + $m)
```

_Note_: dividing by 0 is a runtime error.

# New functionalities (feature flags)

### New function: `overdraft :: (account, asset) -> monetary`

> flag: `experimental-overdraft-function` (available from 0.0.15)

Returns the account's overdraft amount as a positive value (or zero if the account didn't have a negative overdraft)

An example use case is to remove the debt of a certain account:

```numscript
vars {
  monetary $amt = overdraft(@user:001, EUR/2)
}

// I want to remove the debt of @user:001
// if @user:001 has a non negative balance, it will be a noop
/// otherwise we'll bring the @user:001 account to 0
send [COIN *] (
  source = @world
  destination = max $amt to @user:001
)
```

### New function: `get_asset :: monetary -> asset`

> flag: `experimental-get-asset-function` (available from 0.0.16)

Get the asset of the given monetary. For example:

```numscript
vars {
  monetary $mon = [USD/2 100]
  asset $a = get_asset($mon) // => USD/2
}
```

### New function: `get_amount :: monetary -> number`

> flag: `experimental-get-amount-function` (available from 0.0.16)

Get the amount of the given monetary. For example:

```numscript
vars {
  monetary $mon = [USD/2 100]
  number $n = get_amount($mon) // => 100
}
```

### Account interpolation syntax

> flag: `experimental-account-interpolation` (available from 0.0.15)

You can now interpolate variables inside account literals:

```numscript
vars { number $id }

// this will evaluate to e.g. @user:42:pending
@user:$id:pending
```

The interpolation casts implicitly to string the interpolated value.
Only the following types are accepted: `account`, `number`, `string` (interpolating other types will raise a runtime error)
Creating invalid account names (e.g. by interpolating string like `"!"`) will raise a runtime error.

### Mid-script function call

> flag: `experimental-mid-script-function-call` (available from 0.0.15)

The values that initiate vars can now be any kind of expression, not just function calls:

```numscript
vars {
  number $minutes
  number $seconds = 60 * $minutes // <- you can now use any expression here
}
```

At the same time, function can be called mid-script instead of having to define their value in the vars block:

```numscript
send [USD/2 *] (
  source = max balance(@alice, USD/2) from @bob
  destination = @world
)
```

Note that this is more flexible than being forced of defining them in the vars block:

```numscript
send [USD/2 100] (
  source = {
    @acc
    @alice
  }
  destination = @world
)

// if you were to call balance(@alice, USD/2) in the vars block,
// it would return a different value in this case
// as now the balance of alice is lowered by some amount due to the previous statement
send [USD/2 *] (
  source = max balance(@alice, USD/2) from @bob
  destination = @world
)
```

### `oneof` modifier

> flag: `experimental-oneof`

You can add a `oneof` modifier to the inorder blocks (in both source and destination position) so that only the first branch that succeeds is picked. Like the default inorder syntax, it can be nested inside other constructs.

In source position, the first branch that is able to allocate enough funds is picked.
For example:

| account | asset   | amount |
| ------- | ------- | ------ |
| `@a`    | `USD/2` | 99     |
| `@b`    | `USD/2` | 100    |

```numscript
send [USD/2 100] (
  source = oneof {
    @a
    @b
  }
  destination = @dest
)
```

Will produce these postings:

| source | destination | asset   | amount |
| ------ | ----------- | ------- | ------ |
| `@b`   | `@dest`     | `USD/2` | 200    |

As you can see, unlike the default inorder, only a branch is picked, instead of pulling as much as possible from each account.

You can also combine the two syntaxes:

```numscript
// pull either from @a or @b
// but if none of them has enough balance on its own, use the combined balance
send [USD/2 100] (
  source = oneof {
    @a
    @b
    { @a @b }
  }
  destination = @dest
)
```

This also works in destination position: (note that the inorder syntax in destination position has mandatory "max" clauses)

```numscript
send [USD/2 100] (
  source = @world
  destination = oneof {
    max [USD/2 20] to @alice
    max [USD/2 99] to @bob
    max [USD/2 101] to @charlie
    remaining kept
  }
)
```

Will produce:

| source   | destination | asset   | amount |
| -------- | ----------- | ------- | ------ |
| `@world` | `@charlie`  | `USD/2` | 100    |

This may be useful in contexts where the sent amount is injected via a variable, or when the `oneof` destination block is nested within other blocks.

For example, we can top-up the debt (overdraft) of some accounts, by making sure we either fill it or don't send anything:

```numscript
// we'll also need the following for this example:
// experimental-overdraft-function
// experimental-mid-script-function-call
send [USD/2 150] (
  source = @world
  destination = oneof {
    // we prioritize acme which is cheap
    max overdraft(@world:acme) to @world:acme
    // and eventually evilcorp which is expensive
    max overdraft(@world:evilcorp) to @world:evilcorp
    remaining kept
  }
)
```

### Colored assets

> flag: `experimental-asset-colors` (available from 0.0.17)

This functionality allows to deal with semi-fungible assets. While this is already possible, by using conventions like `JPMUSD` and `STRIPEUSD`, there is no way for a statement to deal simultaneously with two different assets.
We therefore introduce a restriction operator that you can use on account on source positions to specify what sub-asset to pull from the balance. An asset `ASSET/n` marked with the "X" color will be represented in the store (for example, the ledger's database) as the `ASSET_X/n` asset.

In practice, the operator looks like this:

```
send [USD/2 100] (
  source = @alice \ "STRIPE"
  destination = @dest
)
```

This will emit the following postings (by checking `@alice`'s `USD_STRIPE/2` balance):

```
[
  {
    source: "alice",
    destination: "dest",
    asset: "USD_STRIPE/2",
    amount: 100,
  }
]
```

A restricted account can nested as usual:

```
send [USD/2 100] (
  source = oneof {
    @alice \ "STRIPE"
    @alice \ "PAYPAL"
    @alice \ "ADYEN"
  }
  destination = @dest
)
```

Colors are represented as string, therefore you use any expression that evaluates to string, including variables:

```
vars {
  string $col
}

send [USD/2 100] (
  source = @alice \ $col
  destination = @dest
)
```

The empty string (`""`) represents no color. Therefore, those two sources are exactly the same:

- `@account \ ""`
- `@account`

In that case, we'll not remap the asset by using the `_` postfix.
