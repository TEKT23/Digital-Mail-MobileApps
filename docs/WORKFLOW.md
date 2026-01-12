# Document Workflows

This document describes the lifecycle and workflows for different types of letters within the Sistem Penyuratan Digital.

## 1. Workflow Surat Keluar (Outgoing Letter)

This workflow involves three main roles: **Staf**, **Manajer**, and **Direktur**. It is designed to ensure that every outgoing letter is properly drafted, verified, and approved before being finalized.

### Roles & Responsibilities

-   **Staf:** Responsible for drafting the initial letter, making revisions, and archiving the final version.
-   **Manajer:** Responsible for verifying the content and format of the letter drafted by the Staf.
-   **Direktur:** Responsible for giving the final approval for the letter.

### Status Lifecycle

The status of an outgoing letter changes as it moves through the workflow:

1.  **`DRAFT`**
    -   A Staf creates a new letter. It is in a draft state and can be edited freely.
    -   **Action:** Staf submits the letter for verification.

2.  **`PERLU_VERIFIKASI`**
    -   The letter is now in the Manajer's queue, awaiting verification.
    -   **Action (Manajer):**
        -   **Approve:** The letter is verified and moves to the next stage.
        -   **Reject:** The letter needs changes and is sent back to the Staf.

3.  **`PERLU_REVISI`** (from Verification or Approval)
    -   The letter has been rejected by either the Manajer or Direktur and is returned to the Staf for modification.
    -   **Action (Staf):** After making revisions, the Staf resubmits the letter, and its status returns to `PERLU_VERIFIKASI`.

4.  **`PERLU_PERSETUJUAN`**
    -   The letter has been verified by the Manajer and is now in the Direktur's queue for final approval.
    -   **Action (Direktur):**
        -   **Approve:** The letter is officially approved.
        -   **Reject:** The letter needs changes and is sent back to the Staf with a status of `PERLU_REVISI`.

5.  **`DISETUJUI`**
    -   The letter has received final approval from the Direktur.
    -   **Action (Staf):** The Staf archives the letter.

6.  **`DIARSIPKAN`**
    -   The workflow is complete. The letter is now a final record and cannot be edited.

---

## 2. Workflow Surat Masuk (Incoming Letter)

This workflow is simpler and involves two main roles: **Staf** and **Direktur**. It is designed to efficiently register and process all incoming letters.

### Roles & Responsibilities

-   **Staf:** Responsible for registering new incoming letters and archiving them after disposition.
-   **Direktur:** Responsible for reviewing incoming letters and providing disposition instructions.

### Status Lifecycle

1.  **`BELUM_DISPOSISI`**
    -   A Staf registers a new incoming letter. It is now in the Direktur's queue, awaiting disposition.
    -   **Action:** The Direktur reviews the letter and adds a disposition.

2.  **`SUDAH_DISPOSISI`**
    -   The Direktur has provided instructions for the letter.
    -   **Action:** The Staf archives the letter.

3.  **`DIARSIPKAN`**
    -   The workflow is complete. The letter and its disposition are saved as a final record.
