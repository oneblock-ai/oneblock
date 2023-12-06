# OneBlock Enhancement Proposal (OEP)

This directory serves as a repository for OneBlock Enhancement Proposals (OEPs). OEPs serve to delineate significant features, enhancements, or bug fixes that require substantial changes to the code or architecture. To create an OEP, you can utilize the provided [template](./YYYYMMDD-template.md) and submit a pull request.

The following steps outline how OEPs play a role in the development process:

1. An engineer or contributor is assigned a GitHub issue labeled `require/OEP`. If the issue lacks this label initially but, upon investigation, the assigned engineer or contributor deems it necessary, they can add the label themselves or reach out to the project's maintainer for label inclusion.
1. Upon deciding to address the issue (usually targeting the next minor release), the assignee initiates work on the OEP. This process may include creating a proof of concept and engaging in extensive discussions.
1. Once the assignee has a preliminary design, they submit a pull request containing the OEP to the `oneblock-ai/oneblock` GitHub repository's `enhancements` directory.
1. The OEP undergoes further discussion, allowing for feedback from the community or team.
1. Immediate merging of the OEP is not required. The assignee might opt to further implement the feature before updating the OEP. Generally, it is necessary to merge the OEP when the feature itself is merged. The OEP ensures that the feature work reflects the initial iteration of effort.
