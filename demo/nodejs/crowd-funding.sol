pragma solidity ^0.4.11;

contract CrowdFunding {
    // Defines a new type with two fields.
    struct Funder {
        address addr;
        uint amount;
    }

    struct Campaign {
        address beneficiary;
        uint fundingGoal;
        uint numFunders;
        uint amount;
        mapping (uint => Funder) funders;
    }

    Campaign campaign;

    event NewContribution(
        address beneficiary,
        address funder,
        uint amount
    );

    function CrowdFunding(address beneficiary, uint goal) {
        // Creates new struct and saves in storage. We leave out the mapping type.
        campaign = Campaign(beneficiary, goal, 0, 0);
    }

    function contribute() payable {
        campaign.funders[campaign.numFunders++] = Funder({addr: msg.sender, amount: msg.value});
        campaign.amount += msg.value;
        NewContribution(campaign.beneficiary, msg.sender, msg.value);
    }

    function checkGoalReached() returns (bool reached) {
        if (campaign.amount < campaign.fundingGoal)
            return false;
        uint amount = campaign.amount;
        campaign.amount = 0;
        campaign.beneficiary.transfer(amount);
        return true;
    }
}