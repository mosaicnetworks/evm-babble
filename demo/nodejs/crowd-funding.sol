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

    function CrowdFunding(uint goal) {
        // Creates new struct and saves in storage. We leave out the mapping type.
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

    function checkGoalReached() returns (bool reached, address beneficiary, uint goal, uint amount) {
        if (campaign.beneficiary == address(0)) {
            return (false, address(0), 0, 0);
        }
        if (campaign.amount < campaign.fundingGoal)
            return (false, campaign.beneficiary,0 , 0);
        uint am = campaign.amount;
        campaign.amount = 0;
        campaign.beneficiary.transfer(am);
        return (true, campaign.beneficiary, campaign.fundingGoal, am);
    }
}