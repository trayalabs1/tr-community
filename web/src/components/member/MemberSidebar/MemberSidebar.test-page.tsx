/**
 * Test Page for Member Sidebar Component
 *
 * To use this for testing:
 * 1. Copy the content to a page file (e.g., app/test/member-sidebar/page.tsx)
 * 2. Navigate to http://localhost:3000/test/member-sidebar
 * 3. Test the component interactively
 *
 * This file shows:
 * - How to import and use the component
 * - How to pass mock data
 * - How to structure the layout
 * - How to use it in a real page context
 */

"use client";

import type { Account } from "@/api/openapi-schema";
import { HStack } from "@/styled-system/jsx";
import { MemberSidebar } from "./MemberSidebar";

// Mock account data
const mockAccount = {
  id: "user-123",
  handle: "raman",
  name: "Raman",
  email: "raman@example.com",
  roles: [
    {
      id: "role-1",
      name: "Member",
    },
  ],
  createdAt: new Date(Date.now() - 42 * 24 * 60 * 60 * 1000).toISOString(),
  updatedAt: new Date().toISOString(),
} as unknown as Account;

// Mock journey stage
const mockJourneyStage = {
  id: "cohort-2",
  name: "Month 2 Cohort",
  memberCount: 312,
};

// Mock topics
const mockTopics = [
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

export function MemberSidebarTestPage() {
  return (
    <HStack
      w="full"
      minH="screen"
      alignItems="start"
      style={{ background: "#f5f5f5" }}
    >
      {/* Sidebar */}
      <div
        style={{
          width: "320px",
          minHeight: "100vh",
          boxShadow: "0 2px 8px rgba(0, 0, 0, 0.1)",
          overflowY: "auto",
        }}
      >
        <MemberSidebar
          account={mockAccount}
          journeyStage={mockJourneyStage}
          topics={mockTopics}
        />
      </div>

      {/* Main Content */}
      <div
        style={{
          flex: 1,
          padding: "40px",
          maxWidth: "1200px",
        }}
      >
        <h1>Member Sidebar Component Test</h1>

        <section style={{ marginBottom: "40px" }}>
          <h2>Component Features</h2>
          <ul>
            <li>✅ Member header with avatar and status</li>
            <li>✅ Bookmark and notification action buttons</li>
            <li>✅ Collapsible "Your Journey Stage" section</li>
            <li>✅ Collapsible "Topics" section</li>
            <li>✅ Smooth hover effects and transitions</li>
            <li>✅ Responsive design</li>
            <li>✅ Traya brand colors applied</li>
            <li>✅ Nunito Sans typography</li>
          </ul>
        </section>

        <section style={{ marginBottom: "40px" }}>
          <h2>Color Palette Used</h2>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: "20px" }}>
            <div>
              <div
                style={{
                  width: "100%",
                  height: "80px",
                  background: "#329866",
                  borderRadius: "8px",
                  marginBottom: "10px",
                }}
              />
              <p>
                <strong>Primary</strong>
                <br />
                #329866
              </p>
            </div>
            <div>
              <div
                style={{
                  width: "100%",
                  height: "80px",
                  background: "#B2E6CD",
                  borderRadius: "8px",
                  marginBottom: "10px",
                  border: "1px solid #ccc",
                }}
              />
              <p>
                <strong>Secondary</strong>
                <br />
                #B2E6CD
              </p>
            </div>
            <div>
              <div
                style={{
                  width: "100%",
                  height: "80px",
                  background: "#DCF4E8",
                  borderRadius: "8px",
                  marginBottom: "10px",
                  border: "1px solid #ccc",
                }}
              />
              <p>
                <strong>Tertiary</strong>
                <br />
                #DCF4E8
              </p>
            </div>
            <div>
              <div
                style={{
                  width: "100%",
                  height: "80px",
                  background: "linear-gradient(180deg, #3EBC7F 17.69%, #329866 100%)",
                  borderRadius: "8px",
                  marginBottom: "10px",
                }}
              />
              <p>
                <strong>Gradient</strong>
                <br />
                #3EBC7F → #329866
              </p>
            </div>
            <div>
              <div
                style={{
                  width: "100%",
                  height: "80px",
                  background: "#F04343",
                  borderRadius: "8px",
                  marginBottom: "10px",
                }}
              />
              <p>
                <strong>Heart</strong>
                <br />
                #F04343
              </p>
            </div>
          </div>
        </section>

        <section style={{ marginBottom: "40px" }}>
          <h2>Testing Checklist</h2>
          <ul>
            <li>☐ Sidebar displays correctly on desktop</li>
            <li>☐ Member header shows avatar with initials</li>
            <li>☐ Bookmark and notification buttons are visible</li>
            <li>☐ Hover on buttons changes to tertiary color</li>
            <li>☐ "Your Journey Stage" section is collapsible</li>
            <li>☐ "Topics" section is collapsible</li>
            <li>☐ Chevron icon rotates on expand/collapse</li>
            <li>☐ Topic cards have hover effects</li>
            <li>☐ Colors match design specification</li>
            <li>☐ Font is Nunito Sans</li>
            <li>☐ Responsive on tablet (iPad size)</li>
            <li>☐ Responsive on mobile (phone size)</li>
          </ul>
        </section>

        <section>
          <h2>Code Example</h2>
          <pre
            style={{
              background: "#f4f4f4",
              padding: "20px",
              borderRadius: "8px",
              overflow: "auto",
            }}
          >
            {`import { MemberSidebar } from "@/components/member/MemberSidebar";

export function MyPage() {
  return (
    <MemberSidebar
      account={currentUser}
      journeyStage={{
        id: "cohort-2",
        name: "Month 2 Cohort",
        memberCount: 312,
      }}
      topics={[
        {
          id: "topic-1",
          name: "Minoxidil Support",
          icon: "💧",
          memberCount: 1243,
        },
        // ... more topics
      ]}
    />
  );
}`}
          </pre>
        </section>
      </div>
    </HStack>
  );
}

export default MemberSidebarTestPage;
