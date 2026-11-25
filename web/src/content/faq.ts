// FAQ data organized by categories
export interface FAQItem {
  id: string;
  question: string;
  answer: string;
  category: string;
}

export const FAQ_DATA: FAQItem[] = [
  // Getting Started
  {
    id: "1",
    question: "How do I get started with the community?",
    category: "Getting Started",
    answer: "After joining, you'll be automatically added to your cohort channel based on your order history. You can explore topic channels, introduce yourself, and start engaging with posts from other members.",
  },
  {
    id: "2",
    question: "What is a cohort channel?",
    category: "Getting Started",
    answer: "Cohort channels group members based on their journey stage (Month 1, Month 2, etc.). You'll find people at similar stages, making it easier to share experiences and support each other.",
  },
  {
    id: "3",
    question: "How do I change my community name?",
    category: "Getting Started",
    answer: "Currently, your community name is set during onboarding. If you need to change it, please contact support.",
  },
  // Using the Community
  {
    id: "5",
    question: "How do I save posts for later?",
    category: "Using the Community",
    answer: "Tap the bookmark icon on any post to save it. Access your saved posts anytime from the bookmark icon in the header.",
  },
  {
    id: "6",
    question: "Can I post images?",
    category: "Using the Community",
    answer: "Yes! When creating a post, you can add images to share your progress or illustrate your questions. Tap the image icon in the post composer.",
  },
  // Guidelines
  {
    id: "7",
    question: "What are the community guidelines?",
    category: "Guidelines",
    answer: "Be respectful and supportive of fellow members. Share your experiences honestly. Avoid medical advice - leave that to our experts. No spam or promotional content. Report any inappropriate behavior.",
  },
  {
    id: "8",
    question: "What happens if I break the rules?",
    category: "Guidelines",
    answer: "Minor violations result in a warning. Repeated or serious violations may lead to temporary or permanent suspension from the community.",
  },
];

export const FAQ_CATEGORIES = ["All", "Getting Started", "Using the Community", "Guidelines"];
