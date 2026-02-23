# Growth projector scenarios

These changes will make the growth projector more impactful by showing different scenarios side-by-side.

## Form updates

### Combine one-time deposit/withdrawal field

Combine the "One-time extra deposit" and "One-time withdrawal" fields into one "One time" field that has an amount input with a selector for deposit/withdrawal. Be not allowing both of these mutually exclusive fields, the number of scenario types is reduced.

### Add a weekly savings field, combined with weekly spending

Similar to the combined deposit/withdrawal field, the "Weekly spending" field should change to "Weekly" and allow the user to select either spending or saving.

## Multiple Scenarios

The growth projector graph will by default will show two lines, each graphing a different scenario. This will allow users to compare the results of different scenarios easily.

Scenarios will show up in the "WHAT IF..." card. Each scenario will be in a row, and a plain English description of the scenario will accompany it as its title. Refer to the "Scenario titles" section to identify how to generate the scenario titles. These titles will be shown in the graph lines' pop-up box above date and balance.

Each scenario will have a distinct color associated with it, which will be the color used for its line on the graph.

Each scenario may be deleted, but there must be at least one scenario.

The user may add additional scenarios.

Keep scenario data in the query so that scenario sets may be bookmarked. When scenarios are added/removed or their data is changed, update the query parameters.

Remove the existing English scenario description.

## Default scenarios

If the user arrives at the growth projector tool without any scenarios in the query parameter, defaul to two scenarios:

### If the child has an allowance

**First scenario**: Weekly spending = 0

**Second scenario**: Weekly spending = 100% of allowance (calculated weekly amount)

### If the child has no allowance

**First scenario**: Weekly spending = 0

**Second scenario**: Weekly spending = $5

## Scenario titles

### Child has an allowance and no one-time deposit or withdrawal

Weekly spending is 0:
> If I save **all** of my $X allowance

Weekly spending == allowance (calculated weekly amount)
> If I save **none** of my $X allowance

0 < weekly spending < allowance (calculated weekly amount)
> If I save **\$Y** per week from my $X allowance

Weekly savings > 0
> If I save **all** of my $X allowance plus an additional **\$Y** per week

### Child has an allowance and a one-time deposit

Weekly spending is 0:
> If I save **all** of my $X allowance, and deposit \$Z now

Weekly spending == allowance (calculated wekly amount)
> If I save **none** of my $X allowance, but deposit \$Z now

0 < weekly spending < allowance (calculated wekly amount)
> If I save **\$Y** per week from my $X allowance, and deposit \$Z now

Weekly savings > 0
> If I save **all** of my $X allowance, plus an additional **\$Y** per week, and deposit **\$Z** now

### Child has an allowance and a one-time withdrawal

Weekly spending is 0:
> If I save **all** of my $X allowance, but withdraw **\$Z** now

Weekly spending == allowance (calculated weekly amount)
> If I save **none** of my $X allowance, and withdraw **\$Z** now

0 < weekly spending < allowance (calculated wekly amount)
> If I save **\$Y** per week from my $X allowance, but withdraw **\$Z** now

Weekly savings > 0
> If I save **all** of my $X allowance, plus an additional **\$Y** per week, but withdraw **\$Z** now

### Child has no allowance and no one-time deposit or withdrawal

Weekly spending is 0:
> If I don't do anything

Weekly spending > 0:
> If I spend **\$X** per week

Weekly savings > 0:
> If I save **\$X** per week

### Child has no allowance and a one-time deposit

Weekly spending is 0:
> If I save **\$X** now

Weekly spending > 0:
> If I save **\$X** now and spend $X per week

### Child has no allowance and a one-time withdrawal

Weekly spending is 0:
> If I spend **\$Z** now

Weekly spending > 0:
> If I spend **\$X** per week, plus withdraw **\$Z** now

Weekly savings > 0:
> If I save **\$X** per week, but withdraw **\$Y** now