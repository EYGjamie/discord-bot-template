export interface Role {
  id: string;
  name: string;
  color: string;
  position: number;
}

export interface MemberInfo {
  user_id: string;
  username: string;
  discriminator: string;
  avatar?: string;
  nickname?: string;
  display_name: string;
}

export interface RolesAndMembersResponse {
  roles: Role[];
  members: MemberInfo[];
}
