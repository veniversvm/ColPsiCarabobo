export interface PostGrade {
  id: string;
  post_grade_title: string;
  post_grade_university: string;
  post_grade_graduation_year: string;
  post_grade_description?: string;
  pic_one_url?: string;
  pic_two_url?: string;
  pic_three_url?: string;
  is_active: boolean;
}