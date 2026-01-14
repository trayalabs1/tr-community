import { MemberSidebar } from "./MemberSidebar";
import { Account } from "@/api/openapi-schema";

/**
 * Example usage of the MemberSidebar component
 *
 * This component displays member information in a sidebar with:
 * - Member header (avatar, name, role, activity status)
 * - Notification and bookmark icons
 * - Collapsible Journey Stage section
 * - Collapsible Topics section
 *
 * Colors used are defined in the component and sourced from:
 * - Primary: #329866
 * - Secondary: #B2E6CD
 * - Tertiary: #DCF4E8
 * - Gradient: linear-gradient(180deg, #3EBC7F 17.69%, #329866 100%)
 * - Heart: #F04343
 *
 * Font: Nunito Sans
 */

// Example data
const exampleAccount: Account = {
  id: "user-123",
  handle: "raman",
  name: "Raman",
  email: "raman@example.com",
  roles: [{ id: "member-role", name: "Member" }],
};

const exampleJourneyStage = {
  id: "cohort-2",
  name: "Month 2 Cohort",
  memberCount: 312,
};

const exampleTopics = [
  {
    id: "topic-1",
    name: "Minoxidil Support",
    icon: "💧",
    memberCount: 1243,
  },
  {
    id: "topic-2",
    name: "Hormones & PCOS",
    icon: "🧬",
    memberCount: 567,
  },
  {
    id: "topic-3",
    name: "Nutrition & Diet",
    icon: "🥗",
    memberCount: 892,
  },
  {
    id: "topic-4",
    name: "Progress Updates",
    icon: "📊",
    memberCount: 1567,
  },
  {
    id: "topic-5",
    name: "Results Gallery",
    icon: "🎨",
    memberCount: 2134,
  },
  {
    id: "topic-6",
    name: "Yoga & Exercise",
    icon: "🧘",
    memberCount: 445,
  },
];

export function MemberSidebarExample() {
  return (
    <MemberSidebar
      account={exampleAccount}
      journeyStage={exampleJourneyStage}
      topics={exampleTopics}
    />
  );
}

/**
 * Integration Guide:
 *
 * 1. Import the component:
 *    import { MemberSidebar } from "@/components/member/MemberSidebar";
 *
 * 2. Get account data from your data source (session, API, etc.)
 * 3. Fetch journey stage and topics if available
 * 4. Render the component:
 *
 *    <MemberSidebar
 *      account={currentUser}
 *      journeyStage={userJourneyStage}
 *      topics={availableTopics}
 *    />
 *
 * Props:
 * - account (required): Account object with user information
 * - journeyStage (optional): Journey stage object with name and member count
 * - topics (optional): Array of topic objects with id, name, icon, and memberCount
 */
