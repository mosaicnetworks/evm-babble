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
    }

    Campaign campaign;

    event NewContribution(
        address beneficiary,
        address funder,
        uint amount
    );

    event Settlement(
        bool ok
    );

    function CrowdFunding(uint goal) {
        // Creates new struct and saves in storage.
        campaign = Campaign({
            beneficiary: msg.sender,
            fundingGoal: goal,
            numFunders: 0,
            amount:0});
    }

    function contribute() payable {
        campaign.amount += msg.value;
        NewContribution(campaign.beneficiary, msg.sender, msg.value);
    }

    function checkGoalReached() constant returns (bool reached, address beneficiary, uint goal, uint amount) {
        if (campaign.amount < campaign.fundingGoal)
            return (false, campaign.beneficiary, campaign.fundingGoal , campaign.amount);
        else 
            return (true, campaign.beneficiary, campaign.fundingGoal , campaign.amount);
    }

    function settle() {
        if (campaign.amount >= campaign.fundingGoal) {
            uint am = campaign.amount;
            campaign.amount = 0;
            campaign.beneficiary.transfer(am);
            Settlement(true);
        } else {
            Settlement(false);
        }
    }
}