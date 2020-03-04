package consensus

var devnet = `[
  {
    "id": 0,
    "desc": "Length in blocks of a voting cycle",
    "type": 0,
    "value": 180
  },
  {
    "id": 1,
    "desc": "Minimum of support needed for starting a range consultation",
    "type": 1,
    "value": 150
  },
  {
    "id": 2,
    "desc": "Minimum of support needed for a consultation answer proposal",
    "type": 1,
    "value": 150
  },
  {
    "id": 3,
    "desc": "Earliest cycle when a consultation can get in confirmation phase",
    "type": 0,
    "value": 2
  },
  {
    "id": 4,
    "desc": "Length in cycles for consultation votings",
    "type": 0,
    "value": 4
  },
  {
    "id": 5,
    "desc": "Maximum of voting cycles for a consultation to gain support",
    "type": 0,
    "value": 4
  },
  {
    "id": 6,
    "desc": "Length in cycles for the reflection phase of consultations",
    "type": 0,
    "value": 1
  },
  {
    "id": 7,
    "desc": "Minimum fee to submit a consultation",
    "type": 2,
    "value": 10000000000
  },
  {
    "id": 8,
    "desc": "Minimum fee to submit a consultation answer proposal",
    "type": 2,
    "value": 5000000000
  },
  {
    "id": 9,
    "desc": "Minimum of quorum for fund proposal votings",
    "type": 1,
    "value": 5000
  },
  {
    "id": 10,
    "desc": "Minimum of positive votes for a fund proposal to be accepted",
    "type": 1,
    "value": 7000
  },
  {
    "id": 11,
    "desc": "Minimum of negative votes for a fund proposal to be rejected",
    "type": 1,
    "value": 7000
  },
  {
    "id": 12,
    "desc": "Minimum fee to submit a fund proposal",
    "type": 2,
    "value": 10000
  },
  {
    "id": 13,
    "desc": "Maximum of voting cycles for fund proposal votings",
    "type": 0,
    "value": 6
  },
  {
    "id": 14,
    "desc": "Minimum of quorum for payment request votings",
    "type": 1,
    "value": 5000
  },
  {
    "id": 15,
    "desc": "Minimum of positive votes for a payment request to be accepted",
    "type": 1,
    "value": 7000
  },
  {
    "id": 16,
    "desc": "Minimum of negative votes for a payment request to be rejected",
    "type": 1,
    "value": 7000
  },
  {
    "id": 17,
    "desc": "Minimum fee to submit a payment request",
    "type": 2,
    "value": 0
  },
  {
    "id": 18,
    "desc": "Maximum of voting cycles for fund proposal votings",
    "type": 0,
    "value": 8
  },
  {
    "id": 19,
    "desc": "Frequency of the fund accumulation transaction",
    "type": 0,
    "value": 500
  },
  {
    "id": 20,
    "desc": "Percentage of generated NAV going to the Fund",
    "type": 1,
    "value": 2000
  },
  {
    "id": 21,
    "desc": "Amount of NAV generated per block",
    "type": 2,
    "value": 250000000
  },
  {
    "id": 22,
    "desc": "Yearly fee for registering a name in NavNS",
    "type": 2,
    "value": 10000000000
  },
  {
    "id": 23,
    "desc": "Minimum fee as a fund contribution to submit a DAO vote using a light wallet",
    "type": 2,
    "value": 10000000
  }
]`
